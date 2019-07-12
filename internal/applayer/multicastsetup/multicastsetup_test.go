package multicastsetup

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
	"github.com/brocaar/lorawan/applayer/multicastsetup"
	"github.com/brocaar/lorawan/gps"
)

type MulticastSetupTestSuite struct {
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
	MulticastGroup   storage.MulticastGroup
}

func (ts *MulticastSetupTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	syncInterval = time.Minute
	syncRetries = 5
	syncBatchSize = 10

	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))
	test.MustResetDB(storage.DB().DB)
}

func (ts *MulticastSetupTestSuite) TearDownTest() {
	ts.tx.Rollback()
}

func (ts *MulticastSetupTestSuite) SetupTest() {
	assert := require.New(ts.T())
	var err error
	ts.tx, err = storage.DB().Beginx()
	assert.NoError(err)

	ts.NSClient = nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(ts.NSClient))

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

	ts.MulticastGroup = storage.MulticastGroup{
		Name:             "test-mg",
		MCAppSKey:        lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		MCKey:            lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateMulticastGroup(ts.tx, &ts.MulticastGroup))
}

func (ts *MulticastSetupTestSuite) TestSyncRemoteMulticastSetupReq() {
	assert := require.New(ts.T())
	ms := storage.RemoteMulticastSetup{
		DevEUI:         ts.Device.DevEUI,
		McGroupID:      1,
		McAddr:         lorawan.DevAddr{1, 2, 3, 4},
		McKeyEncrypted: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		MinMcFCnt:      10,
		MaxMcFCnt:      20,
		State:          storage.RemoteMulticastSetupSetup,
		RetryInterval:  time.Minute,
	}
	copy(ms.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)

	assert.NoError(storage.CreateRemoteMulticastSetup(ts.tx, &ms))
	assert.NoError(syncRemoteMulticastSetup(ts.tx))

	ms, err := storage.GetRemoteMulticastSetup(ts.tx, ms.DevEUI, ms.MulticastGroupID, false)
	assert.NoError(err)
	assert.Equal(1, ms.RetryCount)
	assert.True(ms.RetryAfter.After(time.Now()))

	req := <-ts.NSClient.CreateDeviceQueueItemChan
	assert.Equal(multicastsetup.DefaultFPort, uint8(req.Item.FPort))

	b, err := lorawan.EncryptFRMPayload(ts.DeviceActivation.AppSKey, false, ts.DeviceActivation.DevAddr, 0, req.Item.FrmPayload)
	assert.NoError(err)

	var cmd multicastsetup.Command
	assert.NoError(cmd.UnmarshalBinary(false, b))

	assert.Equal(multicastsetup.Command{
		CID: multicastsetup.McGroupSetupReq,
		Payload: &multicastsetup.McGroupSetupReqPayload{
			McGroupIDHeader: multicastsetup.McGroupSetupReqPayloadMcGroupIDHeader{
				McGroupID: 1,
			},
			McAddr:         ms.McAddr,
			McKeyEncrypted: ms.McKeyEncrypted,
			MinMcFCnt:      ms.MinMcFCnt,
			MaxMcFCnt:      ms.MaxMcFCnt,
		},
	}, cmd)
}

func (ts *MulticastSetupTestSuite) TestMcGroupSetupAns() {
	assert := require.New(ts.T())

	rms := storage.RemoteMulticastSetup{
		DevEUI:         ts.Device.DevEUI,
		McGroupID:      1,
		McAddr:         lorawan.DevAddr{1, 2, 3, 4},
		McKeyEncrypted: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		MinMcFCnt:      10,
		MaxMcFCnt:      20,
		State:          storage.RemoteMulticastSetupSetup,
		RetryInterval:  time.Minute,
	}
	copy(rms.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)
	assert.NoError(storage.CreateRemoteMulticastSetup(ts.tx, &rms))

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		cmd := multicastsetup.Command{
			CID: multicastsetup.McGroupSetupAns,
			Payload: &multicastsetup.McGroupSetupAnsPayload{
				McGroupIDHeader: multicastsetup.McGroupSetupAnsPayloadMcGroupIDHeader{
					IDError:   true,
					McGroupID: 1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.Equal("handle McGroupSetupAns error: IDError for McGroupID: 1", HandleRemoteMulticastSetupCommand(ts.tx, ts.Device.DevEUI, b).Error())
	})

	ts.T().Run("OK", func(t *testing.T) {
		assert := require.New(t)

		cmd := multicastsetup.Command{
			CID: multicastsetup.McGroupSetupAns,
			Payload: &multicastsetup.McGroupSetupAnsPayload{
				McGroupIDHeader: multicastsetup.McGroupSetupAnsPayloadMcGroupIDHeader{
					McGroupID: 1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.NoError(HandleRemoteMulticastSetupCommand(ts.tx, ts.Device.DevEUI, b))

		rms, err := storage.GetRemoteMulticastSetupByGroupID(ts.tx, ts.Device.DevEUI, 1, false)
		assert.NoError(err)
		assert.True(rms.StateProvisioned)
	})
}

func (ts *MulticastSetupTestSuite) TestMcGroupDeleteAns() {
	assert := require.New(ts.T())

	rms := storage.RemoteMulticastSetup{
		DevEUI:         ts.Device.DevEUI,
		McGroupID:      1,
		McAddr:         lorawan.DevAddr{1, 2, 3, 4},
		McKeyEncrypted: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		MinMcFCnt:      10,
		MaxMcFCnt:      20,
		State:          storage.RemoteMulticastSetupDelete,
	}
	copy(rms.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)
	assert.NoError(storage.CreateRemoteMulticastSetup(ts.tx, &rms))
	assert.NoError(storage.AddDeviceToMulticastGroup(ts.tx, rms.MulticastGroupID, ts.Device.DevEUI))

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		cmd := multicastsetup.Command{
			CID: multicastsetup.McGroupDeleteAns,
			Payload: &multicastsetup.McGroupDeleteAnsPayload{
				McGroupIDHeader: multicastsetup.McGroupDeleteAnsPayloadMcGroupIDHeader{
					McGroupUndefined: true,
					McGroupID:        1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.Equal("handle McGroupDeleteAns error: McGroupUndefined for McGroupID: 1", HandleRemoteMulticastSetupCommand(ts.tx, ts.Device.DevEUI, b).Error())

		devices, err := storage.GetDevicesForMulticastGroup(ts.tx, rms.MulticastGroupID, 10, 0)
		assert.NoError(err)
		assert.Len(devices, 1)
	})

	ts.T().Run("OK", func(t *testing.T) {
		assert := require.New(t)

		cmd := multicastsetup.Command{
			CID: multicastsetup.McGroupDeleteAns,
			Payload: &multicastsetup.McGroupDeleteAnsPayload{
				McGroupIDHeader: multicastsetup.McGroupDeleteAnsPayloadMcGroupIDHeader{
					McGroupID: 1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.NoError(HandleRemoteMulticastSetupCommand(ts.tx, ts.Device.DevEUI, b))

		rms, err := storage.GetRemoteMulticastSetupByGroupID(ts.tx, ts.Device.DevEUI, 1, false)
		assert.NoError(err)
		assert.True(rms.StateProvisioned)

		devices, err := storage.GetDevicesForMulticastGroup(ts.tx, rms.MulticastGroupID, 10, 0)
		assert.NoError(err)
		assert.Len(devices, 0)
	})
}

func (ts *MulticastSetupTestSuite) TestSyncRemoteMulticastDeleteReq() {
	assert := require.New(ts.T())

	ms := storage.RemoteMulticastSetup{
		DevEUI:         ts.Device.DevEUI,
		McGroupID:      1,
		McAddr:         lorawan.DevAddr{1, 2, 3, 4},
		McKeyEncrypted: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		MinMcFCnt:      10,
		MaxMcFCnt:      20,
		State:          storage.RemoteMulticastSetupDelete,
		RetryInterval:  time.Minute,
	}
	copy(ms.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)

	assert.NoError(storage.CreateRemoteMulticastSetup(ts.tx, &ms))
	assert.NoError(syncRemoteMulticastSetup(ts.tx))

	ms, err := storage.GetRemoteMulticastSetup(ts.tx, ms.DevEUI, ms.MulticastGroupID, false)
	assert.NoError(err)
	assert.Equal(1, ms.RetryCount)
	assert.True(ms.RetryAfter.After(time.Now()))

	req := <-ts.NSClient.CreateDeviceQueueItemChan
	assert.Equal(multicastsetup.DefaultFPort, uint8(req.Item.FPort))

	b, err := lorawan.EncryptFRMPayload(ts.DeviceActivation.AppSKey, false, ts.DeviceActivation.DevAddr, 0, req.Item.FrmPayload)
	assert.NoError(err)

	var cmd multicastsetup.Command
	assert.NoError(cmd.UnmarshalBinary(false, b))

	assert.Equal(multicastsetup.Command{
		CID: multicastsetup.McGroupDeleteReq,
		Payload: &multicastsetup.McGroupDeleteReqPayload{
			McGroupIDHeader: multicastsetup.McGroupDeleteReqPayloadMcGroupIDHeader{
				McGroupID: 1,
			},
		},
	}, cmd)
}

func (ts *MulticastSetupTestSuite) TestSyncRemoteMulticastClassCSessionReq() {
	assert := require.New(ts.T())
	now := time.Now().Round(time.Second)

	ms := storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		McGroupID:        1,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: true,
	}
	copy(ms.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)
	assert.NoError(storage.CreateRemoteMulticastSetup(ts.tx, &ms))

	sess := storage.RemoteMulticastClassCSession{
		DevEUI:         ts.Device.DevEUI,
		McGroupID:      1,
		SessionTime:    now,
		SessionTimeOut: 10,
		DLFrequency:    868100000,
		DR:             3,
		RetryInterval:  time.Minute,
	}
	copy(sess.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)
	assert.NoError(storage.CreateRemoteMulticastClassCSession(ts.tx, &sess))
	assert.NoError(syncRemoteMulticastClassCSession(ts.tx))

	sess, err := storage.GetRemoteMulticastClassCSession(ts.tx, sess.DevEUI, sess.MulticastGroupID, false)
	assert.NoError(err)
	assert.Equal(1, sess.RetryCount)
	assert.True(sess.RetryAfter.After(time.Now()))

	req := <-ts.NSClient.CreateDeviceQueueItemChan
	assert.Equal(multicastsetup.DefaultFPort, uint8(req.Item.FPort))

	b, err := lorawan.EncryptFRMPayload(ts.DeviceActivation.AppSKey, false, ts.DeviceActivation.DevAddr, 0, req.Item.FrmPayload)
	assert.NoError(err)

	var cmd multicastsetup.Command
	assert.NoError(cmd.UnmarshalBinary(false, b))

	assert.Equal(multicastsetup.Command{
		CID: multicastsetup.McClassCSessionReq,
		Payload: &multicastsetup.McClassCSessionReqPayload{
			McGroupIDHeader: multicastsetup.McClassCSessionReqPayloadMcGroupIDHeader{
				McGroupID: 1,
			},
			SessionTime: uint32((gps.Time(now).TimeSinceGPSEpoch() / time.Second) % (1 << 32)),
			SessionTimeOut: multicastsetup.McClassCSessionReqPayloadSessionTimeOut{
				TimeOut: 10,
			},
			DLFrequency: 868100000,
			DR:          3,
		},
	}, cmd)
}

func (ts *MulticastSetupTestSuite) TestSyncRemoteMulticastClassCSessionAns() {
	assert := require.New(ts.T())

	sess := storage.RemoteMulticastClassCSession{
		DevEUI:         ts.Device.DevEUI,
		McGroupID:      1,
		SessionTimeOut: 10,
		DLFrequency:    868100000,
		DR:             3,
	}
	copy(sess.MulticastGroupID[:], ts.MulticastGroup.MulticastGroup.Id)
	assert.NoError(storage.CreateRemoteMulticastClassCSession(ts.tx, &sess))

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		cmd := multicastsetup.Command{
			CID: multicastsetup.McClassCSessionAns,
			Payload: &multicastsetup.McClassCSessionAnsPayload{
				StatusAndMcGroupID: multicastsetup.McClassCSessionAnsPayloadStatusAndMcGroupID{
					McGroupUndefined: true,
					McGroupID:        1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.Equal("handle McClassCSessionAns error: DRError: false, FreqError: false, McGroupUndefined: true for McGroupID: 1", HandleRemoteMulticastSetupCommand(ts.tx, ts.Device.DevEUI, b).Error())

		devices, err := storage.GetDevicesForMulticastGroup(ts.tx, sess.MulticastGroupID, 10, 0)
		assert.NoError(err)
		assert.Len(devices, 0)
	})

	ts.T().Run("OK", func(t *testing.T) {
		assert := require.New(t)
		tts := uint32(100)

		cmd := multicastsetup.Command{
			CID: multicastsetup.McClassCSessionAns,
			Payload: &multicastsetup.McClassCSessionAnsPayload{
				StatusAndMcGroupID: multicastsetup.McClassCSessionAnsPayloadStatusAndMcGroupID{
					McGroupID: 1,
				},
				TimeToStart: &tts,
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.NoError(HandleRemoteMulticastSetupCommand(ts.tx, ts.Device.DevEUI, b))

		sess, err := storage.GetRemoteMulticastClassCSessionByGroupID(ts.tx, ts.Device.DevEUI, 1, false)
		assert.NoError(err)
		assert.True(sess.StateProvisioned)

		devices, err := storage.GetDevicesForMulticastGroup(ts.tx, sess.MulticastGroupID, 10, 0)
		assert.NoError(err)
		assert.Len(devices, 1)
	})
}

func TestMulticastSetup(t *testing.T) {
	suite.Run(t, new(MulticastSetupTestSuite))
}
