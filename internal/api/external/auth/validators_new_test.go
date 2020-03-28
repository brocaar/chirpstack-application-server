package auth

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type ValidatorTestSuite struct {
	suite.Suite

	networkServers []storage.NetworkServer
	organizations  []storage.Organization
}

func (ts *ValidatorTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	conf := test.GetConfig()
	assert.NoError(storage.Setup(conf))

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))
}

func (ts *ValidatorTestSuite) SetupTest() {
	assert := require.New(ts.T())

	test.MustResetDB(storage.DB().DB)

	ts.networkServers = []storage.NetworkServer{
		{Name: "test-ns", Server: "test-ns:1234"},
		{Name: "test-ns-2", Server: "test-ns-2:1234"},
	}
	for i := range ts.networkServers {
		assert.NoError(storage.CreateNetworkServer(context.Background(), storage.DB(), &ts.networkServers[i]))
	}

	ts.organizations = []storage.Organization{
		{Name: "organization-1", CanHaveGateways: true},
		{Name: "organization-2", CanHaveGateways: false},
	}
	for i := range ts.organizations {
		assert.NoError(storage.CreateOrganization(context.Background(), storage.DB(), &ts.organizations[i]))
	}
}

func (ts *ValidatorTestSuite) CreateUser(username string, isActive, isAdmin bool) (int64, error) {
	u := storage.User{
		Username: username,
		IsAdmin:  isAdmin,
		IsActive: isActive,
		Email:    username + "@example.com",
	}

	return storage.CreateUser(context.Background(), storage.DB(), &u, "v3rys3cr3t!")
}

func (ts *ValidatorTestSuite) RunTests(t *testing.T, tests []validatorTest) {
	for _, tst := range tests {
		t.Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			for _, v := range tst.Validators {
				ok, err := v(storage.DB(), &tst.Claims)
				assert.NoError(err)
				assert.Equal(tst.ExpectedOK, ok)
			}
		})
	}
}

func (ts *ValidatorTestSuite) TestGateway() {
	assert := require.New(ts.T())

	users := []struct {
		username string
		isActive bool
		isAdmin  bool
	}{
		{username: "activeAdmin", isActive: true, isAdmin: true},
		{username: "inactiveAdmin", isActive: false, isAdmin: true},
		{username: "activeUser", isActive: true, isAdmin: false},
		{username: "inactiveUser", isActive: false, isAdmin: false},
	}

	for _, user := range users {
		_, err := ts.CreateUser(user.username, user.isActive, user.isAdmin)
		assert.NoError(err)
	}

	orgUsers := []struct {
		organizationID int64
		username       string
		isAdmin        bool
		isDeviceAdmin  bool
		isGatewayAdmin bool
	}{
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
	}

	for _, orgUser := range orgUsers {
		id, err := ts.CreateUser(orgUser.username, true, false)
		assert.NoError(err)

		err = storage.CreateOrganizationUser(context.Background(), storage.DB(), orgUser.organizationID, id, orgUser.isAdmin, orgUser.isDeviceAdmin, orgUser.isGatewayAdmin)
		assert.NoError(err)
	}

	ts.T().Run("GatewaysAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can create and list",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[0].ID), ValidateGatewaysAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can create and list (org CanHaveGateways=true)",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[0].ID), ValidateGatewaysAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "gateway admin users can create and list (org CanHaveGateways=true)",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[0].ID), ValidateGatewaysAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserGatewayAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "normal user can list",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(List, 0)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can not create (org CanHaveGateways=false)",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[1].ID)},
				Claims:     Claims{Username: "org1ActiveUserAdmin"},
				ExpectedOK: false,
			},
			{
				Name:       "gateway admin users can not create (org CanHaveGateways=true)",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[1].ID)},
				Claims:     Claims{Username: "org1ActiveUserGatewayAdmin"},
				ExpectedOK: false,
			},
			{
				Name:       "organization user can not create",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "normal user can not create",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, ts.organizations[0].ID)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
			{
				Name:       "inactive user can not list",
				Validators: []ValidatorFunc{ValidateGatewaysAccess(List, 0)},
				Claims:     Claims{Username: "inactiveUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})

	ts.T().Run("TestGatewayAccess", func(t *testing.T) {
		assert := require.New(t)

		gateways := []storage.Gateway{
			{MAC: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1}, Name: "gateway1", OrganizationID: ts.organizations[0].ID, NetworkServerID: ts.networkServers[0].ID},
		}
		for i := range gateways {
			assert.NoError(storage.CreateGateway(context.Background(), storage.DB(), &gateways[i]))
		}

		tests := []validatorTest{
			{
				Name:       "global admin users can create, update and delete",
				Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can create, update and delete",
				Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization gateway admin users can create, update and delete",
				Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
				Claims:     Claims{Username: "org0ActiveUserGatewayAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can read",
				Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can not update or delete",
				Validators: []ValidatorFunc{ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "normal users can not read, update or delete",
				Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})
}

func (ts *ValidatorTestSuite) TestApplication() {
	assert := require.New(ts.T())

	users := []struct {
		username string
		isActive bool
		isAdmin  bool
	}{
		{username: "activeAdmin", isActive: true, isAdmin: true},
		{username: "inactiveAdmin", isActive: false, isAdmin: true},
		{username: "activeUser", isActive: true, isAdmin: false},
		{username: "inactiveUser", isActive: false, isAdmin: false},
	}

	for _, user := range users {
		_, err := ts.CreateUser(user.username, user.isActive, user.isAdmin)
		assert.NoError(err)
	}

	orgUsers := []struct {
		organizationID int64
		username       string
		isAdmin        bool
		isDeviceAdmin  bool
		isGatewayAdmin bool
	}{
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
	}

	for _, orgUser := range orgUsers {
		id, err := ts.CreateUser(orgUser.username, true, false)
		assert.NoError(err)

		err = storage.CreateOrganizationUser(context.Background(), storage.DB(), orgUser.organizationID, id, orgUser.isAdmin, orgUser.isDeviceAdmin, orgUser.isGatewayAdmin)
		assert.NoError(err)
	}

	var serviceProfileIDs []uuid.UUID
	serviceProfiles := []storage.ServiceProfile{
		{Name: "test-sp-1", NetworkServerID: ts.networkServers[0].ID, OrganizationID: ts.organizations[0].ID},
	}
	for i := range serviceProfiles {
		assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &serviceProfiles[i]))
		id, _ := uuid.FromBytes(serviceProfiles[i].ServiceProfile.Id)
		serviceProfileIDs = append(serviceProfileIDs, id)
	}

	ts.T().Run("ApplicationsAcccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can create and list",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(Create, ts.organizations[0].ID), ValidateApplicationsAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can create and list",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(Create, ts.organizations[0].ID), ValidateApplicationsAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization device admin users can create and list",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(Create, ts.organizations[0].ID), ValidateApplicationsAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserDeviceAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can list",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "normal users can list when no organization id is given",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(List, 0)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can not create",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(Create, ts.organizations[0].ID)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "normal users can not create and list",
				Validators: []ValidatorFunc{ValidateApplicationsAccess(Create, ts.organizations[0].ID), ValidateApplicationsAccess(List, ts.organizations[0].ID)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})

	ts.T().Run("ApplicationAccess", func(t *testing.T) {
		applications := []storage.Application{
			{OrganizationID: ts.organizations[0].ID, Name: "application-1", ServiceProfileID: serviceProfileIDs[0]},
		}
		for i := range applications {
			assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &applications[i]))
		}

		tests := []validatorTest{
			{
				Name:       "global admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateApplicationAccess(applications[0].ID, Read), ValidateApplicationAccess(applications[0].ID, Update), ValidateApplicationAccess(applications[0].ID, Delete)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can read update and delete",
				Validators: []ValidatorFunc{ValidateApplicationAccess(applications[0].ID, Read), ValidateApplicationAccess(applications[0].ID, Update), ValidateApplicationAccess(applications[0].ID, Delete)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization device admin users can read update and delete",
				Validators: []ValidatorFunc{ValidateApplicationAccess(applications[0].ID, Read), ValidateApplicationAccess(applications[0].ID, Update), ValidateApplicationAccess(applications[0].ID, Delete)},
				Claims:     Claims{Username: "org0ActiveUserDeviceAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can read",
				Validators: []ValidatorFunc{ValidateApplicationAccess(applications[0].ID, Read)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "other users can not read, update or delete",
				Validators: []ValidatorFunc{ValidateApplicationAccess(1, Read), ValidateApplicationAccess(1, Update), ValidateApplicationAccess(1, Delete)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})
}

func (ts *ValidatorTestSuite) TestDevice() {
	assert := require.New(ts.T())

	users := []struct {
		username string
		isActive bool
		isAdmin  bool
	}{
		{username: "activeAdmin", isActive: true, isAdmin: true},
		{username: "inactiveAdmin", isActive: false, isAdmin: true},
		{username: "activeUser", isActive: true, isAdmin: false},
		{username: "inactiveUser", isActive: false, isAdmin: false},
	}

	for _, user := range users {
		_, err := ts.CreateUser(user.username, user.isActive, user.isAdmin)
		assert.NoError(err)
	}

	orgUsers := []struct {
		organizationID int64
		username       string
		isAdmin        bool
		isDeviceAdmin  bool
		isGatewayAdmin bool
	}{
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
	}

	for _, orgUser := range orgUsers {
		id, err := ts.CreateUser(orgUser.username, true, false)
		assert.NoError(err)

		err = storage.CreateOrganizationUser(context.Background(), storage.DB(), orgUser.organizationID, id, orgUser.isAdmin, orgUser.isDeviceAdmin, orgUser.isGatewayAdmin)
		assert.NoError(err)
	}

	var serviceProfileIDs []uuid.UUID
	serviceProfiles := []storage.ServiceProfile{
		{Name: "test-sp-1", NetworkServerID: ts.networkServers[0].ID, OrganizationID: ts.organizations[0].ID},
	}
	for i := range serviceProfiles {
		assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &serviceProfiles[i]))
		id, _ := uuid.FromBytes(serviceProfiles[i].ServiceProfile.Id)
		serviceProfileIDs = append(serviceProfileIDs, id)
	}

	applications := []storage.Application{
		{OrganizationID: ts.organizations[0].ID, Name: "application-1", ServiceProfileID: serviceProfileIDs[0]},
	}
	for i := range applications {
		assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &applications[i]))
	}

	deviceProfiles := []storage.DeviceProfile{
		{Name: "test-dp-1", OrganizationID: ts.organizations[0].ID, NetworkServerID: ts.networkServers[0].ID},
	}
	var deviceProfilesIDs []uuid.UUID
	for i := range deviceProfiles {
		assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &deviceProfiles[i]))
		dpID, _ := uuid.FromBytes(deviceProfiles[i].DeviceProfile.Id)
		deviceProfilesIDs = append(deviceProfilesIDs, dpID)
	}

	devices := []storage.Device{
		{DevEUI: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1}, Name: "test-1", ApplicationID: applications[0].ID, DeviceProfileID: deviceProfilesIDs[0]},
	}
	for i := range devices {
		assert.NoError(storage.CreateDevice(context.Background(), storage.DB(), &devices[i]))
	}

	ts.T().Run("DevicesAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin user has access to create and list",
				Validators: []ValidatorFunc{ValidateNodesAccess(applications[0].ID, Create), ValidateNodesAccess(applications[0].ID, List)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can create and list",
				Validators: []ValidatorFunc{ValidateNodesAccess(applications[0].ID, Create), ValidateNodesAccess(applications[0].ID, List)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization device admin users can create and list",
				Validators: []ValidatorFunc{ValidateNodesAccess(applications[0].ID, Create), ValidateNodesAccess(applications[0].ID, List)},
				Claims:     Claims{Username: "org0ActiveUserDeviceAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can list",
				Validators: []ValidatorFunc{ValidateNodesAccess(applications[0].ID, List)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can not create",
				Validators: []ValidatorFunc{ValidateNodesAccess(applications[0].ID, Create)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "other users can not create or list",
				Validators: []ValidatorFunc{ValidateNodesAccess(applications[0].ID, Create), ValidateNodesAccess(applications[0].ID, List)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})

	ts.T().Run("DeviceAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateNodeAccess(devices[0].DevEUI, Read), ValidateNodeAccess(devices[0].DevEUI, Update), ValidateNodeAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateNodeAccess(devices[0].DevEUI, Read), ValidateNodeAccess(devices[0].DevEUI, Update), ValidateNodeAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization device admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateNodeAccess(devices[0].DevEUI, Read), ValidateNodeAccess(devices[0].DevEUI, Update), ValidateNodeAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "org0ActiveUserDeviceAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can read",
				Validators: []ValidatorFunc{ValidateNodeAccess(devices[0].DevEUI, Read)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users (non-admin) can not update or delete",
				Validators: []ValidatorFunc{ValidateNodeAccess(devices[0].DevEUI, Update), ValidateNodeAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "other users can not read, update and delete",
				Validators: []ValidatorFunc{ValidateNodeAccess(devices[0].DevEUI, Read), ValidateNodeAccess(devices[0].DevEUI, Update), ValidateNodeAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})

	ts.T().Run("DeviceQueueAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can read, list, update and delete",
				Validators: []ValidatorFunc{ValidateDeviceQueueAccess(devices[0].DevEUI, Create), ValidateDeviceQueueAccess(devices[0].DevEUI, List), ValidateDeviceQueueAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can read, list, update and delete",
				Validators: []ValidatorFunc{ValidateDeviceQueueAccess(devices[0].DevEUI, Create), ValidateDeviceQueueAccess(devices[0].DevEUI, List), ValidateDeviceQueueAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "other users can not read, list, update and delete",
				Validators: []ValidatorFunc{ValidateDeviceQueueAccess(devices[0].DevEUI, Create), ValidateDeviceQueueAccess(devices[0].DevEUI, List), ValidateDeviceQueueAccess(devices[0].DevEUI, Delete)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})
}

func (ts *ValidatorTestSuite) TestDeviceProfile() {
	assert := require.New(ts.T())

	users := []struct {
		username string
		isActive bool
		isAdmin  bool
	}{
		{username: "activeAdmin", isActive: true, isAdmin: true},
		{username: "inactiveAdmin", isActive: false, isAdmin: true},
		{username: "activeUser", isActive: true, isAdmin: false},
		{username: "inactiveUser", isActive: false, isAdmin: false},
	}

	for _, user := range users {
		_, err := ts.CreateUser(user.username, user.isActive, user.isAdmin)
		assert.NoError(err)
	}

	orgUsers := []struct {
		organizationID int64
		username       string
		isAdmin        bool
		isDeviceAdmin  bool
		isGatewayAdmin bool
	}{
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
	}

	for _, orgUser := range orgUsers {
		id, err := ts.CreateUser(orgUser.username, true, false)
		assert.NoError(err)

		err = storage.CreateOrganizationUser(context.Background(), storage.DB(), orgUser.organizationID, id, orgUser.isAdmin, orgUser.isDeviceAdmin, orgUser.isGatewayAdmin)
		assert.NoError(err)
	}

	var serviceProfileIDs []uuid.UUID
	serviceProfiles := []storage.ServiceProfile{
		{Name: "test-sp-1", NetworkServerID: ts.networkServers[0].ID, OrganizationID: ts.organizations[0].ID},
	}
	for i := range serviceProfiles {
		assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &serviceProfiles[i]))
		id, _ := uuid.FromBytes(serviceProfiles[i].ServiceProfile.Id)
		serviceProfileIDs = append(serviceProfileIDs, id)
	}

	applications := []storage.Application{
		{OrganizationID: ts.organizations[0].ID, Name: "application-1", ServiceProfileID: serviceProfileIDs[0]},
		{OrganizationID: ts.organizations[1].ID, Name: "application-2", ServiceProfileID: serviceProfileIDs[0]},
	}
	for i := range applications {
		assert.NoError(storage.CreateApplication(context.Background(), storage.DB(), &applications[i]))
	}

	deviceProfiles := []storage.DeviceProfile{
		{Name: "test-dp-1", OrganizationID: ts.organizations[0].ID, NetworkServerID: ts.networkServers[0].ID},
	}
	var deviceProfilesIDs []uuid.UUID
	for i := range deviceProfiles {
		assert.NoError(storage.CreateDeviceProfile(context.Background(), storage.DB(), &deviceProfiles[i]))
		dpID, _ := uuid.FromBytes(deviceProfiles[i].DeviceProfile.Id)
		deviceProfilesIDs = append(deviceProfilesIDs, dpID)
	}

	ts.T().Run("DeviceProfilesAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can create and list",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(Create, ts.organizations[0].ID, 0), ValidateDeviceProfilesAccess(List, ts.organizations[0].ID, 0)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can create and list",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(Create, ts.organizations[0].ID, 0), ValidateDeviceProfilesAccess(List, ts.organizations[0].ID, 0)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization device admin users can create and list",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(Create, ts.organizations[0].ID, 0), ValidateDeviceProfilesAccess(List, ts.organizations[0].ID, 0)},
				Claims:     Claims{Username: "org0ActiveUserDeviceAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can list",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(List, ts.organizations[0].ID, 0)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can list with an application id is given",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(List, 0, applications[0].ID)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "any user can list when organization id = 0",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(List, 0, 0)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can not create",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(Create, ts.organizations[0].ID, 0)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "non-organization users can not create or list",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(Create, ts.organizations[0].ID, 0), ValidateDeviceProfilesAccess(List, ts.organizations[0].ID, 0)},
				Claims:     Claims{Username: "org1ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "non-organization users can not list when an application id is given beloning to a different organization",
				Validators: []ValidatorFunc{ValidateDeviceProfilesAccess(List, 0, applications[1].ID)},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})

	ts.T().Run("DeviceProfileAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateDeviceProfileAccess(Read, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Update, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Delete, deviceProfilesIDs[0])},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateDeviceProfileAccess(Read, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Update, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Delete, deviceProfilesIDs[0])},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization device admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateDeviceProfileAccess(Read, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Update, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Delete, deviceProfilesIDs[0])},
				Claims:     Claims{Username: "org0ActiveUserDeviceAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can read",
				Validators: []ValidatorFunc{ValidateDeviceProfileAccess(Read, deviceProfilesIDs[0])},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: true,
			},
			{
				Name:       "organization users can not update and delete",
				Validators: []ValidatorFunc{ValidateDeviceProfileAccess(Update, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Delete, deviceProfilesIDs[0])},
				Claims:     Claims{Username: "org0ActiveUser"},
				ExpectedOK: false,
			},
			{
				Name:       "non-organization users can not read, update ande delete",
				Validators: []ValidatorFunc{ValidateDeviceProfileAccess(Read, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Update, deviceProfilesIDs[0]), ValidateDeviceProfileAccess(Delete, deviceProfilesIDs[0])},
				Claims:     Claims{Username: "org1ActiveUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})
}

func (ts *ValidatorTestSuite) TestNetworkServer() {
	assert := require.New(ts.T())

	users := []struct {
		username string
		isActive bool
		isAdmin  bool
	}{
		{username: "activeAdmin", isActive: true, isAdmin: true},
		{username: "inactiveAdmin", isActive: false, isAdmin: true},
		{username: "activeUser", isActive: true, isAdmin: false},
		{username: "inactiveUser", isActive: false, isAdmin: false},
	}

	for _, user := range users {
		_, err := ts.CreateUser(user.username, user.isActive, user.isAdmin)
		assert.NoError(err)
	}

	orgUsers := []struct {
		organizationID int64
		username       string
		isAdmin        bool
		isDeviceAdmin  bool
		isGatewayAdmin bool
	}{
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[0].ID, username: "org0ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUser", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserAdmin", isAdmin: true, isDeviceAdmin: false, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserDeviceAdmin", isAdmin: false, isDeviceAdmin: true, isGatewayAdmin: false},
		{organizationID: ts.organizations[1].ID, username: "org1ActiveUserGatewayAdmin", isAdmin: false, isDeviceAdmin: false, isGatewayAdmin: true},
	}

	for _, orgUser := range orgUsers {
		id, err := ts.CreateUser(orgUser.username, true, false)
		assert.NoError(err)

		err = storage.CreateOrganizationUser(context.Background(), storage.DB(), orgUser.organizationID, id, orgUser.isAdmin, orgUser.isDeviceAdmin, orgUser.isGatewayAdmin)
		assert.NoError(err)
	}

	var serviceProfileIDs []uuid.UUID
	serviceProfiles := []storage.ServiceProfile{
		{Name: "test-sp-1", NetworkServerID: ts.networkServers[0].ID, OrganizationID: ts.organizations[0].ID},
	}
	for i := range serviceProfiles {
		assert.NoError(storage.CreateServiceProfile(context.Background(), storage.DB(), &serviceProfiles[i]))
		id, _ := uuid.FromBytes(serviceProfiles[i].ServiceProfile.Id)
		serviceProfileIDs = append(serviceProfileIDs, id)
	}

	ts.T().Run("NetworkServerAccess", func(t *testing.T) {
		tests := []validatorTest{
			{
				Name:       "global admin users can read, update and delete",
				Validators: []ValidatorFunc{ValidateNetworkServerAccess(Read, ts.networkServers[0].ID), ValidateNetworkServerAccess(Update, ts.networkServers[0].ID), ValidateNetworkServerAccess(Delete, ts.networkServers[0].ID)},
				Claims:     Claims{Username: "activeAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can read",
				Validators: []ValidatorFunc{ValidateNetworkServerAccess(Read, ts.networkServers[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization gateway admin users can read",
				Validators: []ValidatorFunc{ValidateNetworkServerAccess(Read, ts.networkServers[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserGatewayAdmin"},
				ExpectedOK: true,
			},
			{
				Name:       "organization admin users can not update and delete",
				Validators: []ValidatorFunc{ValidateNetworkServerAccess(Update, ts.networkServers[0].ID), ValidateNetworkServerAccess(Delete, ts.networkServers[0].ID)},
				Claims:     Claims{Username: "org0ActiveUserAdmin"},
				ExpectedOK: false,
			},
			{
				Name:       "regular users can not read, update and delete",
				Validators: []ValidatorFunc{ValidateNetworkServerAccess(Read, ts.networkServers[0].ID), ValidateNetworkServerAccess(Update, ts.networkServers[0].ID), ValidateNetworkServerAccess(Delete, ts.networkServers[0].ID)},
				Claims:     Claims{Username: "activeUser"},
				ExpectedOK: false,
			},
		}

		ts.RunTests(t, tests)
	})
}

func TestValidatorsNew(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
