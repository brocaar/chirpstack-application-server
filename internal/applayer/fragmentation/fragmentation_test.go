package fragmentation

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
	"github.com/brocaar/lorawan/applayer/fragmentation"
)

type FragmentationSessionTestSuite struct {
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

func (ts *FragmentationSessionTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	syncInterval = time.Minute
	syncRetries = 5
	syncBatchSize = 10

	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))
	test.MustResetDB(storage.DB().DB)
}

func (ts *FragmentationSessionTestSuite) TearDownTest() {
	assert := require.New(ts.T())
	assert.NoError(ts.tx.Rollback())
}

func (ts *FragmentationSessionTestSuite) SetupTest() {
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
}

func (ts *FragmentationSessionTestSuite) TestSyncFragSessionSetupReq() {
	assert := require.New(ts.T())
	rfs := storage.RemoteFragmentationSession{
		DevEUI:              ts.Device.DevEUI,
		FragIndex:           1,
		NbFrag:              10,
		FragSize:            50,
		FragmentationMatrix: 5,
		BlockAckDelay:       3,
		Padding:             2,
		Descriptor:          [4]byte{1, 2, 3, 4},
		State:               storage.RemoteMulticastSetupSetup,
		RetryInterval:       time.Minute,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(ts.tx, &rfs))
	assert.NoError(syncRemoteFragmentationSessions(ts.tx))

	rfs, err := storage.GetRemoteFragmentationSession(ts.tx, ts.Device.DevEUI, 1, false)
	assert.NoError(err)
	assert.Equal(1, rfs.RetryCount)
	assert.True(rfs.RetryAfter.After(time.Now()))

	req := <-ts.NSClient.CreateDeviceQueueItemChan
	assert.Equal(fragmentation.DefaultFPort, uint8(req.Item.FPort))

	b, err := lorawan.EncryptFRMPayload(ts.DeviceActivation.AppSKey, false, ts.DeviceActivation.DevAddr, 0, req.Item.FrmPayload)
	assert.NoError(err)

	var cmd fragmentation.Command
	assert.NoError(cmd.UnmarshalBinary(false, b))

	assert.Equal(fragmentation.Command{
		CID: fragmentation.FragSessionSetupReq,
		Payload: &fragmentation.FragSessionSetupReqPayload{
			FragSession: fragmentation.FragSessionSetupReqPayloadFragSession{
				FragIndex: 1,
			},
			NbFrag:   10,
			FragSize: 50,
			Control: fragmentation.FragSessionSetupReqPayloadControl{
				FragmentationMatrix: 5,
				BlockAckDelay:       3,
			},
			Padding:    2,
			Descriptor: [4]byte{1, 2, 3, 4},
		},
	}, cmd)
}

func (ts *FragmentationSessionTestSuite) TestSyncFragSessionDeleteReq() {
	assert := require.New(ts.T())

	rfs := storage.RemoteFragmentationSession{
		DevEUI:              ts.Device.DevEUI,
		FragIndex:           1,
		NbFrag:              10,
		FragSize:            50,
		FragmentationMatrix: 5,
		BlockAckDelay:       3,
		Padding:             2,
		Descriptor:          [4]byte{1, 2, 3, 4},
		State:               storage.RemoteMulticastSetupDelete,
		RetryInterval:       time.Minute,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(ts.tx, &rfs))
	assert.NoError(syncRemoteFragmentationSessions(ts.tx))

	rfs, err := storage.GetRemoteFragmentationSession(ts.tx, ts.Device.DevEUI, 1, false)
	assert.NoError(err)
	assert.Equal(1, rfs.RetryCount)
	assert.True(rfs.RetryAfter.After(time.Now()))

	req := <-ts.NSClient.CreateDeviceQueueItemChan
	assert.Equal(fragmentation.DefaultFPort, uint8(req.Item.FPort))

	b, err := lorawan.EncryptFRMPayload(ts.DeviceActivation.AppSKey, false, ts.DeviceActivation.DevAddr, 0, req.Item.FrmPayload)
	assert.NoError(err)

	var cmd fragmentation.Command
	assert.NoError(cmd.UnmarshalBinary(false, b))

	assert.Equal(fragmentation.Command{
		CID: fragmentation.FragSessionDeleteReq,
		Payload: &fragmentation.FragSessionDeleteReqPayload{
			Param: fragmentation.FragSessionDeleteReqPayloadParam{
				FragIndex: 1,
			},
		},
	}, cmd)
}

func (ts *FragmentationSessionTestSuite) TestFragSessionSetupAns() {
	assert := require.New(ts.T())

	rfs := storage.RemoteFragmentationSession{
		DevEUI:    ts.Device.DevEUI,
		FragIndex: 1,
		State:     storage.RemoteMulticastSetupSetup,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(ts.tx, &rfs))

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		cmd := fragmentation.Command{
			CID: fragmentation.FragSessionSetupAns,
			Payload: &fragmentation.FragSessionSetupAnsPayload{
				StatusBitMask: fragmentation.FragSessionSetupAnsPayloadStatusBitMask{
					FragIndex:       1,
					WrongDescriptor: true,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.Equal("handle FragSessionSetupAns error: WrongDescriptor: true, FragSessionIndexNotSupported: false, NotEnoughMemory: false, EncodingUnsupported: false", HandleRemoteFragmentationSessionCommand(ts.tx, ts.Device.DevEUI, b).Error())
	})

	ts.T().Run("OK", func(t *testing.T) {
		assert := require.New(t)

		cmd := fragmentation.Command{
			CID: fragmentation.FragSessionSetupAns,
			Payload: &fragmentation.FragSessionSetupAnsPayload{
				StatusBitMask: fragmentation.FragSessionSetupAnsPayloadStatusBitMask{
					FragIndex: 1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.NoError(HandleRemoteFragmentationSessionCommand(ts.tx, ts.Device.DevEUI, b))

		rfs, err := storage.GetRemoteFragmentationSession(ts.tx, ts.Device.DevEUI, 1, false)
		assert.NoError(err)
		assert.True(rfs.StateProvisioned)
	})
}

func (ts *FragmentationSessionTestSuite) TestFragSessionDeleteAns() {
	assert := require.New(ts.T())

	rfs := storage.RemoteFragmentationSession{
		DevEUI:    ts.Device.DevEUI,
		FragIndex: 1,
		State:     storage.RemoteMulticastSetupSetup,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(ts.tx, &rfs))

	ts.T().Run("Error", func(t *testing.T) {
		assert := require.New(t)

		cmd := fragmentation.Command{
			CID: fragmentation.FragSessionDeleteAns,
			Payload: &fragmentation.FragSessionDeleteAnsPayload{
				Status: fragmentation.FragSessionDeleteAnsPayloadStatus{
					FragIndex:           1,
					SessionDoesNotExist: true,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.Equal("handle FragSessionDeleteAns error: FragIndex 1 does not exist", HandleRemoteFragmentationSessionCommand(ts.tx, ts.Device.DevEUI, b).Error())
	})

	ts.T().Run("OK", func(t *testing.T) {
		assert := require.New(t)

		cmd := fragmentation.Command{
			CID: fragmentation.FragSessionDeleteAns,
			Payload: &fragmentation.FragSessionDeleteAnsPayload{
				Status: fragmentation.FragSessionDeleteAnsPayloadStatus{
					FragIndex: 1,
				},
			},
		}
		b, err := cmd.MarshalBinary()
		assert.NoError(err)
		assert.NoError(HandleRemoteFragmentationSessionCommand(ts.tx, ts.Device.DevEUI, b))

		rfs, err := storage.GetRemoteFragmentationSession(ts.tx, ts.Device.DevEUI, 1, false)
		assert.NoError(err)
		assert.True(rfs.StateProvisioned)
	})
}

func (ts *FragmentationSessionTestSuite) TestFragSessionStatusAns() {
	assert := require.New(ts.T())
	fd := storage.FUOTADeployment{
		Name: "test-deployment",
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(ts.tx, &fd, ts.Device.DevEUI))
	fdd, err := storage.GetPendingFUOTADeploymentDevice(ts.tx, ts.Device.DevEUI)
	assert.NoError(err)

	tests := []struct {
		Name                 string
		FragSessionStatusAns fragmentation.FragSessionStatusAnsPayload
		ExpectedState        storage.FUOTADeploymentDeviceState
		ExpectedErrorMessage string
	}{
		{
			Name: "success",
			FragSessionStatusAns: fragmentation.FragSessionStatusAnsPayload{
				ReceivedAndIndex: fragmentation.FragSessionStatusAnsPayloadReceivedAndIndex{
					FragIndex:      0,
					NbFragReceived: 10,
				},
			},
			ExpectedState: storage.FUOTADeploymentDeviceSuccess,
		},
		{
			Name: "missing frags",
			FragSessionStatusAns: fragmentation.FragSessionStatusAnsPayload{
				ReceivedAndIndex: fragmentation.FragSessionStatusAnsPayloadReceivedAndIndex{
					FragIndex:      0,
					NbFragReceived: 10,
				},
				MissingFrag: 20,
			},
			ExpectedState:        storage.FUOTADeploymentDeviceError,
			ExpectedErrorMessage: "20 fragments missed (10 received).",
		},
		{
			Name: "not enough matrix memory",
			FragSessionStatusAns: fragmentation.FragSessionStatusAnsPayload{
				ReceivedAndIndex: fragmentation.FragSessionStatusAnsPayloadReceivedAndIndex{
					FragIndex:      0,
					NbFragReceived: 10,
				},
				Status: fragmentation.FragSessionStatusAnsPayloadStatus{
					NotEnoughMatrixMemory: true,
				},
			},
			ExpectedState:        storage.FUOTADeploymentDeviceError,
			ExpectedErrorMessage: "Not enough matrix memory.",
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			// start with a clean state
			fdd.State = storage.FUOTADeploymentDevicePending
			fdd.ErrorMessage = ""
			assert.NoError(storage.UpdateFUOTADeploymentDevice(ts.tx, &fdd))

			cmd := fragmentation.Command{
				CID:     fragmentation.FragSessionStatusAns,
				Payload: &tst.FragSessionStatusAns,
			}
			b, err := cmd.MarshalBinary()
			assert.NoError(err)

			assert.NoError(HandleRemoteFragmentationSessionCommand(ts.tx, ts.Device.DevEUI, b))

			devices, err := storage.GetFUOTADeploymentDevices(ts.tx, fd.ID, 10, 0)
			assert.NoError(err)
			assert.Len(devices, 1)
			assert.Equal(tst.ExpectedState, devices[0].State)
			assert.Equal(tst.ExpectedErrorMessage, devices[0].ErrorMessage)
		})
	}
}

func TestFragmentationSession(t *testing.T) {
	suite.Run(t, new(FragmentationSessionTestSuite))
}
