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

func (ts *StorageTestSuite) TestRemoteMulticastClassCSession() {
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

	rms := RemoteMulticastSetup{
		DevEUI:           d.DevEUI,
		MulticastGroupID: mgID,
		McGroupID:        1,
		State:            RemoteMulticastSetupSetup,
		StateProvisioned: false,
	}
	assert.NoError(CreateRemoteMulticastSetup(ts.tx, &rms))

	ts.T().Run("Create", func(t *testing.T) {
		assert := require.New(t)

		sess := RemoteMulticastClassCSession{
			DevEUI:           d.DevEUI,
			MulticastGroupID: mgID,
			McGroupID:        1,
			SessionTime:      now.Add(time.Minute),
			SessionTimeOut:   10,
			DLFrequency:      868100000,
			DR:               3,
			RetryAfter:       now,
			RetryCount:       1,
			RetryInterval:    time.Minute,
		}
		assert.NoError(CreateRemoteMulticastClassCSession(ts.tx, &sess))
		sess.CreatedAt = sess.CreatedAt.UTC().Round(time.Millisecond)
		sess.UpdatedAt = sess.UpdatedAt.UTC().Round(time.Millisecond)
		sess.RetryAfter = sess.RetryAfter.UTC().Round(time.Millisecond)
		sess.SessionTime = sess.SessionTime.UTC().Round(time.Millisecond)

		t.Run("Get", func(t *testing.T) {
			assert := require.New(t)

			sessGet, err := GetRemoteMulticastClassCSession(ts.tx, d.DevEUI, mgID, false)
			assert.NoError(err)
			sessGet.CreatedAt = sessGet.CreatedAt.UTC().Round(time.Millisecond)
			sessGet.UpdatedAt = sessGet.UpdatedAt.UTC().Round(time.Millisecond)
			sessGet.RetryAfter = sessGet.RetryAfter.UTC().Round(time.Millisecond)
			sessGet.SessionTime = sessGet.SessionTime.UTC().Round(time.Millisecond)
			assert.Equal(sess, sessGet)
		})

		t.Run("GetPending no setup", func(t *testing.T) {
			assert := require.New(t)

			items, err := GetPendingRemoteMulticastClassCSessions(ts.tx, 10, 2)
			assert.NoError(err)
			assert.Len(items, 0)
		})

		t.Run("GetPending", func(t *testing.T) {
			assert := require.New(t)

			rms.StateProvisioned = true
			assert.NoError(UpdateRemoteMulticastSetup(ts.tx, &rms))

			items, err := GetPendingRemoteMulticastClassCSessions(ts.tx, 10, 2)
			assert.NoError(err)
			assert.Len(items, 1)

			// start a new transaction and make sure that we do not get the locked
			// items in the result-set.
			newTX, err := DB().Beginx()
			assert.NoError(err)

			items, err = GetPendingRemoteMulticastClassCSessions(newTX, 10, 2)
			assert.NoError(err)
			assert.Len(items, 0)

			assert.NoError(newTX.Rollback())
		})

		t.Run("Update", func(t *testing.T) {
			assert := require.New(t)
			now = now.Add(time.Second)

			sess.McGroupID = 3
			sess.SessionTime = now.Add(time.Hour)
			sess.SessionTimeOut = 20
			sess.DLFrequency = 86300000
			sess.DR = 2
			sess.StateProvisioned = true
			sess.RetryAfter = now
			sess.RetryInterval = time.Minute * 2
			assert.NoError(UpdateRemoteMulticastClassCSession(ts.tx, &sess))
			sess.UpdatedAt = sess.UpdatedAt.UTC().Round(time.Millisecond)

			sessGet, err := GetRemoteMulticastClassCSession(ts.tx, d.DevEUI, mgID, false)
			assert.NoError(err)
			sessGet.CreatedAt = sessGet.CreatedAt.UTC().Round(time.Millisecond)
			sessGet.UpdatedAt = sessGet.UpdatedAt.UTC().Round(time.Millisecond)
			sessGet.RetryAfter = sessGet.RetryAfter.UTC().Round(time.Millisecond)
			sessGet.SessionTime = sessGet.SessionTime.UTC().Round(time.Millisecond)
			assert.Equal(sess, sessGet)

			t.Run("GetPending", func(t *testing.T) {
				assert := require.New(t)

				items, err := GetPendingRemoteMulticastClassCSessions(ts.tx, 10, 2)
				assert.NoError(err)
				assert.Len(items, 0)
			})
		})

		t.Run("Delete", func(t *testing.T) {
			assert := require.New(t)

			assert.NoError(DeleteRemoteMulticastClassCSession(ts.tx, d.DevEUI, mgID))
			_, err := GetRemoteMulticastClassCSession(ts.tx, d.DevEUI, mgID, false)
			assert.Equal(err, ErrDoesNotExist)
		})
	})
}
