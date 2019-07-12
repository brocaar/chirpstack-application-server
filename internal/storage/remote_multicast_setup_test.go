package storage

import (
	"testing"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lorawan"
)

func (ts *StorageTestSuite) TestRemoteMulticastSetup() {
	assert := require.New(ts.T())

	nsClient := nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(nsClient))

	n := NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(CreateNetworkServer(ts.tx, &n))

	org := Organization{
		Name: "test-org",
	}
	assert.NoError(CreateOrganization(ts.tx, &org))

	sp := ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(CreateServiceProfile(ts.tx, &sp))
	var spID uuid.UUID
	copy(spID[:], sp.ServiceProfile.Id)

	app := Application{
		Name:             "test-app",
		OrganizationID:   org.ID,
		ServiceProfileID: spID,
	}
	assert.NoError(CreateApplication(ts.tx, &app))

	dp := DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  org.ID,
		NetworkServerID: n.ID,
	}
	assert.NoError(CreateDeviceProfile(ts.tx, &dp))
	var dpID uuid.UUID
	copy(dpID[:], dp.DeviceProfile.Id)

	d := Device{
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		ApplicationID:   app.ID,
		DeviceProfileID: dpID,
		Name:            "test-device",
		Description:     "test device",
	}
	assert.NoError(CreateDevice(ts.tx, &d))

	mg := MulticastGroup{
		Name:             "test-mg",
		MCAppSKey:        lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		MCKey:            lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		ServiceProfileID: spID,
	}
	assert.NoError(CreateMulticastGroup(ts.tx, &mg))
	var mgID uuid.UUID
	copy(mgID[:], mg.MulticastGroup.Id)

	now := time.Now().UTC().Round(time.Millisecond)

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		dmg := RemoteMulticastSetup{
			DevEUI:           d.DevEUI,
			MulticastGroupID: mgID,
			McGroupID:        2,
			McAddr:           lorawan.DevAddr{1, 2, 3, 4},
			McKeyEncrypted:   lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
			MinMcFCnt:        10,
			MaxMcFCnt:        20,
			State:            RemoteMulticastSetupSetup,
			RetryAfter:       now,
			RetryCount:       1,
			RetryInterval:    time.Minute,
		}
		assert.NoError(CreateRemoteMulticastSetup(ts.tx, &dmg))
		dmg.CreatedAt = dmg.CreatedAt.UTC().Round(time.Millisecond)
		dmg.UpdatedAt = dmg.UpdatedAt.UTC().Round(time.Millisecond)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			dmgGet, err := GetRemoteMulticastSetup(ts.tx, d.DevEUI, mgID, false)
			assert.NoError(err)
			dmgGet.CreatedAt = dmgGet.CreatedAt.UTC().Round(time.Millisecond)
			dmgGet.UpdatedAt = dmgGet.UpdatedAt.UTC().Round(time.Millisecond)
			dmgGet.RetryAfter = dmgGet.RetryAfter.UTC()
			assert.Equal(dmg, dmgGet)
		})

		t.Run("GetPending", func(t *testing.T) {
			assert := require.New(t)

			items, err := GetPendingRemoteMulticastSetupItems(ts.tx, 10, 2)
			assert.NoError(err)
			assert.Len(items, 1)

			// start a new transaction and make sure that we do not get the locked
			// items in the result-set.
			newTX, err := DB().Beginx()
			assert.NoError(err)

			items, err = GetPendingRemoteMulticastSetupItems(newTX, 10, 2)
			assert.NoError(err)
			assert.Len(items, 0)

			assert.NoError(newTX.Rollback())
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)
			now = now.Add(time.Second)

			dmg.McGroupID = 3
			dmg.McAddr = lorawan.DevAddr{4, 3, 2, 1}
			dmg.McKeyEncrypted = lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1}
			dmg.MinMcFCnt = 100
			dmg.MaxMcFCnt = 200
			dmg.State = RemoteMulticastSetupDelete
			dmg.StateProvisioned = true
			dmg.RetryAfter = now
			dmg.RetryInterval = time.Minute * 2
			assert.NoError(UpdateRemoteMulticastSetup(ts.tx, &dmg))
			dmg.UpdatedAt = dmg.UpdatedAt.UTC().Round(time.Millisecond)

			dmgGet, err := GetRemoteMulticastSetup(ts.tx, d.DevEUI, mgID, false)
			assert.NoError(err)
			dmgGet.CreatedAt = dmgGet.CreatedAt.UTC().Round(time.Millisecond)
			dmgGet.UpdatedAt = dmgGet.UpdatedAt.UTC().Round(time.Millisecond)
			dmgGet.RetryAfter = dmgGet.RetryAfter.UTC()
			assert.Equal(dmg, dmgGet)

			t.Run("GetPending", func(t *testing.T) {
				assert := require.New(t)

				items, err := GetPendingRemoteMulticastSetupItems(ts.tx, 10, 2)
				assert.NoError(err)
				assert.Len(items, 0)
			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteRemoteMulticastSetup(ts.tx, d.DevEUI, mgID))
			_, err := GetRemoteMulticastSetup(ts.tx, d.DevEUI, mgID, false)
			assert.Equal(err, ErrDoesNotExist)
		})
	})
}
