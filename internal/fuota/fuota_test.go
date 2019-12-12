package fuota

import (
	"context"
	"testing"
	"time"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	nsmock "github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/applayer/fragmentation"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FUOTATestSuite struct {
	suite.Suite

	tx       *storage.TxLogger
	nsClient *nsmock.Client

	NetworkServer    storage.NetworkServer
	Organization     storage.Organization
	ServiceProfile   storage.ServiceProfile
	Application      storage.Application
	DeviceProfile    storage.DeviceProfile
	Device           storage.Device
	DeviceActivation storage.DeviceActivation
}

func (ts *FUOTATestSuite) SetupSuite() {
	assert := require.New(ts.T())
	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))
	test.MustResetDB(storage.DB().DB)

	remoteMulticastSetupRetries = 3
	remoteFragmentationSessionRetries = 3
}

func (ts *FUOTATestSuite) TearDownTest() {
	ts.tx.Rollback()
}

func (ts *FUOTATestSuite) SetupTest() {
	assert := require.New(ts.T())
	var err error
	ts.tx, err = storage.DB().Beginx()
	assert.NoError(err)

	ts.nsClient = nsmock.NewClient()
	networkserver.SetPool(nsmock.NewPool(ts.nsClient))

	ts.NetworkServer = storage.NetworkServer{
		Name:   "test",
		Server: "test:1234",
	}
	assert.NoError(storage.CreateNetworkServer(context.Background(), ts.tx, &ts.NetworkServer))

	ts.Organization = storage.Organization{
		Name: "test-org",
	}
	assert.NoError(storage.CreateOrganization(context.Background(), ts.tx, &ts.Organization))

	ts.ServiceProfile = storage.ServiceProfile{
		Name:            "test-sp",
		OrganizationID:  ts.Organization.ID,
		NetworkServerID: ts.NetworkServer.ID,
	}
	assert.NoError(storage.CreateServiceProfile(context.Background(), ts.tx, &ts.ServiceProfile))
	var spID uuid.UUID
	copy(spID[:], ts.ServiceProfile.ServiceProfile.Id)

	ts.Application = storage.Application{
		Name:             "test-app",
		OrganizationID:   ts.Organization.ID,
		ServiceProfileID: spID,
	}
	assert.NoError(storage.CreateApplication(context.Background(), ts.tx, &ts.Application))

	ts.DeviceProfile = storage.DeviceProfile{
		Name:            "test-dp",
		OrganizationID:  ts.Organization.ID,
		NetworkServerID: ts.NetworkServer.ID,
	}
	assert.NoError(storage.CreateDeviceProfile(context.Background(), ts.tx, &ts.DeviceProfile))
	var dpID uuid.UUID
	copy(dpID[:], ts.DeviceProfile.DeviceProfile.Id)

	ts.Device = storage.Device{
		DevEUI:          lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		ApplicationID:   ts.Application.ID,
		DeviceProfileID: dpID,
		Name:            "test-device",
		Description:     "test device",
	}
	assert.NoError(storage.CreateDevice(context.Background(), ts.tx, &ts.Device))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentMulticastCreate() {
	assert := require.New(ts.T())

	fd := storage.FUOTADeployment{
		Name:           "test-deployment",
		GroupType:      storage.FUOTADeploymentGroupTypeB,
		DR:             3,
		Frequency:      868100000,
		PingSlotPeriod: 2,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// test that multicast-group has been set
	fdGet, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.NotNil(fdGet.MulticastGroupID)
	assert.Equal(storage.FUOTADeploymentMulticastSetup, fdGet.State)

	// test that multicast-group has been created
	mgCreateReq := <-ts.nsClient.CreateMulticastGroupChan
	assert.NotNil(mgCreateReq.MulticastGroup)
	assert.NotNil(mgCreateReq.MulticastGroup.McAddr)
	assert.NotNil(mgCreateReq.MulticastGroup.McNwkSKey)
	assert.EqualValues(0, mgCreateReq.MulticastGroup.FCnt)
	assert.EqualValues(3, mgCreateReq.MulticastGroup.Dr)
	assert.EqualValues(868100000, mgCreateReq.MulticastGroup.Frequency)
	assert.EqualValues(2, mgCreateReq.MulticastGroup.PingSlotPeriod)
	assert.EqualValues(ts.ServiceProfile.ServiceProfile.Id, mgCreateReq.MulticastGroup.ServiceProfileId)
	assert.EqualValues(routingProfileID.Bytes(), mgCreateReq.MulticastGroup.RoutingProfileId)

	mg, err := storage.GetMulticastGroup(context.Background(), ts.tx, *fdGet.MulticastGroupID, false, true)
	assert.NoError(err)
	assert.NotEqual("", mg.Name)
	assert.NotEqual(lorawan.AES128Key{}, mg.MCAppSKey)
	assert.NotEqual(lorawan.AES128Key{}, mg.MCKey)
	assert.EqualValues(ts.ServiceProfile.ServiceProfile.Id, mg.ServiceProfileID.Bytes())
}

func (ts *FUOTATestSuite) TestFUOTADeploymentMulticastSetupLW10() {
	assert := require.New(ts.T())

	// init
	deviceKeys := storage.DeviceKeys{
		DevEUI:    ts.Device.DevEUI,
		GenAppKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}
	assert.NoError(storage.CreateDeviceKeys(context.Background(), ts.tx, &deviceKeys))

	mcg := storage.MulticastGroup{
		Name:  "test-mg",
		MCKey: lorawan.AES128Key{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)
	mcgReq := <-ts.nsClient.CreateMulticastGroupChan
	ts.nsClient.GetMulticastGroupResponse.MulticastGroup = mcgReq.MulticastGroup

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		UnicastTimeout:   time.Second,
		State:            storage.FUOTADeploymentMulticastSetup,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate remote multicast setup
	items, err := storage.GetPendingRemoteMulticastSetupItems(context.Background(), ts.tx, 10, 10)
	assert.NoError(err)
	assert.Len(items, 1)

	items[0].CreatedAt = time.Time{}
	items[0].UpdatedAt = time.Time{}
	items[0].RetryAfter = time.Time{}

	assert.Equal(storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		MaxMcFCnt:        (1 << 32) - 1,
		McKeyEncrypted:   lorawan.AES128Key{0xba, 0x6a, 0xbb, 0xd4, 0xe4, 0x10, 0xa, 0x62, 0xb9, 0x81, 0xa8, 0x2a, 0xb9, 0x47, 0xd4, 0xa},
		State:            storage.RemoteMulticastSetupSetup,
		RetryInterval:    time.Second,
	}, items[0])

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentFragmentationSessSetup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.After(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentMulticastSetupLW11() {
	assert := require.New(ts.T())

	// init
	deviceKeys := storage.DeviceKeys{
		DevEUI:    ts.Device.DevEUI,
		AppKey:    lorawan.AES128Key{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		GenAppKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}
	assert.NoError(storage.CreateDeviceKeys(context.Background(), ts.tx, &deviceKeys))

	mcg := storage.MulticastGroup{
		Name:  "test-mg",
		MCKey: lorawan.AES128Key{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)
	mcgReq := <-ts.nsClient.CreateMulticastGroupChan
	ts.nsClient.GetMulticastGroupResponse.MulticastGroup = mcgReq.MulticastGroup

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		UnicastTimeout:   time.Second,
		State:            storage.FUOTADeploymentMulticastSetup,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate remote multicast setup
	items, err := storage.GetPendingRemoteMulticastSetupItems(context.Background(), ts.tx, 10, 10)
	assert.NoError(err)
	assert.Len(items, 1)

	items[0].CreatedAt = time.Time{}
	items[0].UpdatedAt = time.Time{}
	items[0].RetryAfter = time.Time{}

	assert.Equal(storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		MaxMcFCnt:        (1 << 32) - 1,
		McKeyEncrypted:   lorawan.AES128Key{0x3a, 0xaa, 0x69, 0xd9, 0x60, 0x6a, 0x9a, 0xdf, 0xad, 0x54, 0x4e, 0x76, 0x5d, 0x5f, 0xe6, 0xd3},
		State:            storage.RemoteMulticastSetupSetup,
		RetryInterval:    time.Second,
	}, items[0])

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentFragmentationSessSetup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.After(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentFragmentationSessionSetup() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name:  "test-mg",
		MCKey: lorawan.AES128Key{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)
	mcgReq := <-ts.nsClient.CreateMulticastGroupChan
	ts.nsClient.GetMulticastGroupResponse.MulticastGroup = mcgReq.MulticastGroup

	fd := storage.FUOTADeployment{
		Name:                "test-deployment",
		MulticastGroupID:    &mcgID,
		UnicastTimeout:      time.Second,
		State:               storage.FUOTADeploymentFragmentationSessSetup,
		FragmentationMatrix: 3,
		Descriptor:          [4]byte{1, 2, 3, 4},
		Payload:             []byte{1, 2, 3, 4, 5},
		FragSize:            2,
		Redundancy:          10,
		BlockAckDelay:       4,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	rms := storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: true,
	}
	assert.NoError(storage.CreateRemoteMulticastSetup(context.Background(), ts.tx, &rms))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate fragmentation sesssion
	items, err := storage.GetPendingRemoteFragmentationSessions(context.Background(), ts.tx, 10, 10)
	assert.NoError(err)
	assert.Len(items, 1)

	items[0].CreatedAt = time.Time{}
	items[0].UpdatedAt = time.Time{}
	items[0].RetryAfter = time.Time{}

	assert.Equal(storage.RemoteFragmentationSession{
		DevEUI:              ts.Device.DevEUI,
		FragIndex:           0,
		MCGroupIDs:          []int{0},
		NbFrag:              3,
		FragSize:            2,
		FragmentationMatrix: 3,
		BlockAckDelay:       4,
		Padding:             1,
		Descriptor:          [4]byte{1, 2, 3, 4},
		State:               storage.RemoteMulticastSetupSetup,
		RetryInterval:       time.Second,
	}, items[0])

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentMulticastSessCSetup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.After(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentFragmentationSessionSetupMulticastSetupNotCompleted() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name:  "test-mg",
		MCKey: lorawan.AES128Key{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)
	mcgReq := <-ts.nsClient.CreateMulticastGroupChan
	ts.nsClient.GetMulticastGroupResponse.MulticastGroup = mcgReq.MulticastGroup

	fd := storage.FUOTADeployment{
		Name:                "test-deployment",
		MulticastGroupID:    &mcgID,
		UnicastTimeout:      time.Second,
		State:               storage.FUOTADeploymentFragmentationSessSetup,
		FragmentationMatrix: 3,
		Descriptor:          [4]byte{1, 2, 3, 4},
		Payload:             []byte{1, 2, 3, 4, 5},
		FragSize:            2,
		Redundancy:          10,
		BlockAckDelay:       4,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	rms := storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: false,
	}
	assert.NoError(storage.CreateRemoteMulticastSetup(context.Background(), ts.tx, &rms))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate fragmentation sesssion
	items, err := storage.GetPendingRemoteFragmentationSessions(context.Background(), ts.tx, 10, 10)
	assert.NoError(err)
	assert.Len(items, 0)
}

func (ts *FUOTATestSuite) TestFUOTADeploymentMulticastSessCSetup() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
		MulticastGroup: ns.MulticastGroup{
			Frequency: 868100000,
			Dr:        5,
		},
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)
	mcgReq := <-ts.nsClient.CreateMulticastGroupChan
	ts.nsClient.GetMulticastGroupResponse.MulticastGroup = mcgReq.MulticastGroup

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		UnicastTimeout:   time.Second,
		MulticastTimeout: 8,
		State:            storage.FUOTADeploymentMulticastSessCSetup,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	rms := storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: true,
	}
	assert.NoError(storage.CreateRemoteMulticastSetup(context.Background(), ts.tx, &rms))

	rfs := storage.RemoteFragmentationSession{
		DevEUI:           ts.Device.DevEUI,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: true,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(context.Background(), ts.tx, &rfs))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate class-c sessions
	items, err := storage.GetPendingRemoteMulticastClassCSessions(context.Background(), ts.tx, 10, 10)
	assert.NoError(err)
	assert.Len(items, 1)

	items[0].CreatedAt = time.Time{}
	items[0].UpdatedAt = time.Time{}
	items[0].RetryAfter = time.Time{}

	assert.True(items[0].SessionTime.After(time.Now()))
	items[0].SessionTime = time.Time{}

	assert.Equal(storage.RemoteMulticastClassCSession{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		DLFrequency:      868100000,
		DR:               5,
		SessionTimeOut:   8,
		RetryInterval:    time.Second,
	}, items[0])
}

func (ts *FUOTATestSuite) TestFUOTADeploymentEnqueue() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name:      "test-mg",
		MCAppSKey: lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		MulticastGroup: ns.MulticastGroup{
			FCnt: 10,
		},
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)
	mcgReq := <-ts.nsClient.CreateMulticastGroupChan
	ts.nsClient.GetMulticastGroupResponse.MulticastGroup = mcgReq.MulticastGroup

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		Payload:          []byte{1, 2, 3, 4},
		FragSize:         2,
		Redundancy:       1,
		State:            storage.FUOTADeploymentEnqueue,
		GroupType:        storage.FUOTADeploymentGroupTypeC,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	// run
	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate scheduled payloads
	items := []ns.MulticastQueueItem{
		{
			MulticastGroupId: mcgID.Bytes(),
			FrmPayload:       []byte{0xe2, 0xfd, 0x27, 0xb0, 0x1b},
			FCnt:             10,
			FPort:            uint32(fragmentation.DefaultFPort),
		},
		{
			MulticastGroupId: mcgID.Bytes(),
			FrmPayload:       []byte{0x60, 0x19, 0x2d, 0x1d, 0x37},
			FCnt:             11,
			FPort:            uint32(fragmentation.DefaultFPort),
		},
		{
			MulticastGroupId: mcgID.Bytes(),
			FrmPayload:       []byte{0x76, 0x30, 0x39, 0xac, 0xae},
			FCnt:             12,
			FPort:            uint32(fragmentation.DefaultFPort),
		},
	}

	for _, item := range items {
		req := <-ts.nsClient.EnqueueMulticastQueueItemChan
		assert.Equal(item, *req.MulticastQueueItem)
	}
}

func (ts *FUOTATestSuite) TestFUOTADeploymentStatusRequest() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)

	rms := storage.RemoteMulticastSetup{
		DevEUI:           ts.Device.DevEUI,
		MulticastGroupID: mcgID,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: true,
	}
	assert.NoError(storage.CreateRemoteMulticastSetup(context.Background(), ts.tx, &rms))

	rfs := storage.RemoteFragmentationSession{
		DevEUI:           ts.Device.DevEUI,
		FragIndex:        fragIndex,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: true,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(context.Background(), ts.tx, &rfs))

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		State:            storage.FUOTADeploymentStatusRequest,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate
	req := <-ts.nsClient.CreateDeviceQueueItemChan
	assert.NotNil(req.Item)
	assert.Equal(ns.DeviceQueueItem{
		DevAddr:    ts.DeviceActivation.DevAddr[:],
		DevEui:     ts.Device.DevEUI[:],
		FrmPayload: []byte{0x73, 0xa5},
		FPort:      uint32(fragmentation.DefaultFPort),
	}, *req.Item)

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentSetDeviceStatus, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.Before(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentSetDeviceStatusNoError() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		State:            storage.FUOTADeploymentSetDeviceStatus,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	fdd, err := storage.GetPendingFUOTADeploymentDevice(context.Background(), ts.tx, ts.Device.DevEUI)
	assert.NoError(err)
	fdd.State = storage.FUOTADeploymentDeviceSuccess
	assert.NoError(storage.UpdateFUOTADeploymentDevice(context.Background(), ts.tx, &fdd))

	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	items, err := storage.GetFUOTADeploymentDevices(context.Background(), ts.tx, fd.ID, 10, 0)
	assert.NoError(err)
	assert.Len(items, 1)
	assert.Equal(storage.FUOTADeploymentDeviceSuccess, items[0].State)
	assert.Equal("", items[0].ErrorMessage)

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentCleanup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.Before(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentSetDeviceStatusGenericError() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		State:            storage.FUOTADeploymentSetDeviceStatus,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	items, err := storage.GetFUOTADeploymentDevices(context.Background(), ts.tx, fd.ID, 10, 0)
	assert.NoError(err)
	assert.Len(items, 1)
	assert.Equal(storage.FUOTADeploymentDeviceError, items[0].State)
	assert.Equal("Device did not complete the FUOTA deployment or did not confirm that it completed the FUOTA deployment.", items[0].ErrorMessage)

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentCleanup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.Before(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentSetDeviceStatusRemoteMulticastSetupError() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)

	rms := storage.RemoteMulticastSetup{
		MulticastGroupID: mcgID,
		DevEUI:           ts.Device.DevEUI,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: false,
	}
	assert.NoError(storage.CreateRemoteMulticastSetup(context.Background(), ts.tx, &rms))

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		State:            storage.FUOTADeploymentSetDeviceStatus,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	items, err := storage.GetFUOTADeploymentDevices(context.Background(), ts.tx, fd.ID, 10, 0)
	assert.NoError(err)
	assert.Len(items, 1)
	assert.Equal(storage.FUOTADeploymentDeviceError, items[0].State)
	assert.Equal("The device failed to provision the remote multicast setup.", items[0].ErrorMessage)

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentCleanup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.Before(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentSetDeviceStatusFragmentationSessionError() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)

	rfs := storage.RemoteFragmentationSession{
		FragIndex:        fragIndex,
		DevEUI:           ts.Device.DevEUI,
		State:            storage.RemoteMulticastSetupSetup,
		StateProvisioned: false,
	}
	assert.NoError(storage.CreateRemoteFragmentationSession(context.Background(), ts.tx, &rfs))

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		State:            storage.FUOTADeploymentSetDeviceStatus,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	items, err := storage.GetFUOTADeploymentDevices(context.Background(), ts.tx, fd.ID, 10, 0)
	assert.NoError(err)
	assert.Len(items, 1)
	assert.Equal(storage.FUOTADeploymentDeviceError, items[0].State)
	assert.Equal("The device failed to provision the fragmentation session setup.", items[0].ErrorMessage)

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentCleanup, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.Before(time.Now()))
}

func (ts *FUOTATestSuite) TestFUOTADeploymentCleanup() {
	assert := require.New(ts.T())

	mcg := storage.MulticastGroup{
		Name: "test-mg",
	}
	copy(mcg.ServiceProfileID[:], ts.ServiceProfile.ServiceProfile.Id)
	assert.NoError(storage.CreateMulticastGroup(context.Background(), ts.tx, &mcg))
	var mcgID uuid.UUID
	copy(mcgID[:], mcg.MulticastGroup.Id)

	fd := storage.FUOTADeployment{
		Name:             "test-deployment",
		MulticastGroupID: &mcgID,
		State:            storage.FUOTADeploymentCleanup,
	}
	assert.NoError(storage.CreateFUOTADeploymentForDevice(context.Background(), ts.tx, &fd, ts.Device.DevEUI))

	assert.NoError(fuotaDeployments(context.Background(), ts.tx))

	// validate fuota deployment record
	fdUpdated, err := storage.GetFUOTADeployment(context.Background(), ts.tx, fd.ID, false)
	assert.NoError(err)
	assert.Equal(storage.FUOTADeploymentDone, fdUpdated.State)
	assert.True(fdUpdated.NextStepAfter.Before(time.Now()))
	assert.Nil(fdUpdated.MulticastGroupID)

	_, err = storage.GetMulticastGroup(context.Background(), ts.tx, mcgID, false, false)
	assert.Equal(storage.ErrDoesNotExist, err)
}

func TestFUOTA(t *testing.T) {
	suite.Run(t, new(FUOTATestSuite))
}
