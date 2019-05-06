package clocksync

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/lora-app-server/internal/backend/networkserver/mock"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/applayer/clocksync"
	"github.com/brocaar/lorawan/gps"
)

type ClockSyncTestSuite struct {
	suite.Suite
	tx *storage.TxLogger

	NSClient         *nsmock.Client
	NetworkServer    storage.NetworkServer
	Organization     storage.Organization
	ServiceProfile   storage.ServiceProfile
	Application      storage.Application
	DeviceProfile    storage.DeviceProfile
	Device           storage.Device
	DeviceActivation storage.DeviceActivation
}

func (ts *ClockSyncTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))
	test.MustResetDB(storage.DB().DB)
}

func (ts *ClockSyncTestSuite) TearDownTest() {
	ts.tx.Rollback()
}

func (ts *ClockSyncTestSuite) SetupTest() {
	assert := require.New(ts.T())

	ts.NSClient = nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(ts.NSClient))

	var err error
	ts.tx, err = storage.DB().Beginx()
	assert.NoError(err)

	ts.NetworkServer = storage.NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(ts.tx, &ts.NetworkServer))

	ts.Organization = storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(ts.tx, &ts.Organization))

	ts.ServiceProfile = storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  ts.Organization.ID,
		NetworkServerID: ts.NetworkServer.ID,
	}
	assert.NoError(storage.CreateServiceProfile(ts.tx, &ts.ServiceProfile))
	var spID uuid.UUID
	copy(spID[:], ts.ServiceProfile.ServiceProfile.Id)

	ts.Application = storage.Application{
		Name:             "test-app",
		OrganizationID:   ts.Organization.ID,
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(ts.tx, &ts.Application))

	ts.DeviceProfile = storage.DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  ts.Organization.ID,
		NetworkServerID: ts.NetworkServer.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(ts.tx, &ts.DeviceProfile))
	var dpID uuid.UUID
	copy(dpID[:], ts.DeviceProfile.DeviceProfile.Id)

	ts.Device = storage.Device{
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		ApplicationID:   ts.Application.ID,
		DeviceProfileID: dpID,
		Name:            "test-device",
		Description:     "test device",
	}
	assert.NoError(storage.CreateDevice(ts.tx, &ts.Device))

	ts.DeviceActivation = storage.DeviceActivation{
		DevEUI: ts.Device.DevEUI,
	}
	assert.NoError(storage.CreateDeviceActivation(ts.tx, &ts.DeviceActivation))
}

func (ts *ClockSyncTestSuite) TestAppTimeReq() {
	deviceTime := time.Now()
	serverTime := deviceTime.Add(20 * time.Second)

	deviceGPSTime := gps.Time(deviceTime).TimeSinceGPSEpoch()
	serverGPSTime := gps.Time(serverTime).TimeSinceGPSEpoch()

	ts.T().Run("AnsRequired", func(t *testing.T) {
		assert := require.New(t)

		cmd := clocksync.Command{
			CID: clocksync.AppTimeReq,
			Payload: &clocksync.AppTimeReqPayload{
				DeviceTime: uint32((deviceGPSTime / time.Second) % (1 << 32)),
				Param: clocksync.AppTimeReqPayloadParam{
					AnsRequired: true,
					TokenReq:    15,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.NoError(HandleClockSyncCommand(ts.tx, ts.Device.DevEUI, serverGPSTime, b))

		queueReq := <-ts.NSClient.CreateDeviceQueueItemChan

		var ans clocksync.Command
		assert.Equal(clocksync.DefaultFPort, uint8(queueReq.Item.FPort))

		b, err = lorawan.EncryptFRMPayload(ts.DeviceActivation.AppSKey, false, ts.DeviceActivation.DevAddr, 0, queueReq.Item.FrmPayload)
		assert.NoError(err)

		assert.NoError(ans.UnmarshalBinary(false, b))
		assert.Equal(clocksync.Command{
			CID: clocksync.AppTimeAns,
			Payload: &clocksync.AppTimeAnsPayload{
				Param: clocksync.AppTimeAnsPayloadParam{
					TokenAns: 15,
				},
				TimeCorrection: 20,
			},
		}, ans)
	})

}

func TestClockSynchronization(t *testing.T) {
	suite.Run(t, new(ClockSyncTestSuite))
}
