package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver/mock"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
	"github.com/brocaar/lorawan"
)

type validatorTest struct {
	Name       string
	Claims     Claims
	Validators []ValidatorFunc
	ExpectedOK bool
}

func TestValidators(t *testing.T) {
	conf := test.GetConfig()
	if err := storage.Setup(conf); err != nil {
		t.Fatal(err)
	}

	test.MustResetDB(storage.DB().DB)

	nsClient := mock.NewClient()
	networkserver.SetPool(mock.NewPool(nsClient))

	/*
	   Users:
	   1: global admin
	   4: no membership
	   8: global admin (but is_active=false)
	   9: member of organization 1
	   10: admin of organization 1
	   11: member of organization 1 (but is_active=false)
	   12: admin of organization 2

	   Organizations:
	   1: organization 1 (can have gateways)
	   2: organization 2 (can not have gateways)

	   Applications:
	   1: application 1
	   2: application 2

	   Nodes:
	   0101010101010101: application 1 node
	   0202020202020202: application 2 node

	   Gateways:
	   0101010101010101: organization 1 gw
	   0202020202020202: organization 2 gw

	   FUOTA deployment:
	   1: created for device 0101010101010101

	*/
	networkServers := []storage.NetworkServer{
		{Name: "test-ns", Server: "test-ns:1234"},
		{Name: "test-ns-2", Server: "test-ns-2:1234"},
	}
	for i := range networkServers {
		if err := storage.CreateNetworkServer(context.Background(), storage.DB(), &networkServers[i]); err != nil {
			t.Fatal(err)
		}
	}

	organizations := []storage.Organization{
		{Name: "organization-1", CanHaveGateways: true},
		{Name: "organization-2", CanHaveGateways: false},
	}
	serviceProfiles := []storage.ServiceProfile{
		{Name: "test-sp-1", NetworkServerID: networkServers[0].ID},
		{Name: "test-sp-2", NetworkServerID: networkServers[0].ID},
	}
	var serviceProfilesIDs []uuid.UUID
	for i := range organizations {
		if err := storage.CreateOrganization(context.Background(), storage.DB(), &organizations[i]); err != nil {
			t.Fatal(err)
		}

		serviceProfiles[i].OrganizationID = organizations[i].ID
		if err := storage.CreateServiceProfile(context.Background(), storage.DB(), &serviceProfiles[i]); err != nil {
			t.Fatal(err)
		}

		spID, _ := uuid.FromBytes(serviceProfiles[i].ServiceProfile.Id)
		serviceProfilesIDs = append(serviceProfilesIDs, spID)
	}

	multicastGroups := []storage.MulticastGroup{
		{Name: "mg-1", ServiceProfileID: serviceProfilesIDs[0]},
		{Name: "mg-2", ServiceProfileID: serviceProfilesIDs[1]},
	}
	var multicastGroupsIDs []uuid.UUID
	for i := range multicastGroups {
		if err := storage.CreateMulticastGroup(context.Background(), storage.DB(), &multicastGroups[i]); err != nil {
			t.Fatal(err)
		}

		mgID, _ := uuid.FromBytes(multicastGroups[i].MulticastGroup.Id)
		multicastGroupsIDs = append(multicastGroupsIDs, mgID)
	}

	deviceProfiles := []storage.DeviceProfile{
		{Name: "test-dp-1", OrganizationID: organizations[0].ID, NetworkServerID: networkServers[0].ID},
		{Name: "test-dp-2", OrganizationID: organizations[1].ID, NetworkServerID: networkServers[0].ID},
	}
	var deviceProfilesIDs []uuid.UUID
	for i := range deviceProfiles {
		if err := storage.CreateDeviceProfile(context.Background(), storage.DB(), &deviceProfiles[i]); err != nil {
			t.Fatal(err)
		}
		dpID, _ := uuid.FromBytes(deviceProfiles[i].DeviceProfile.Id)
		deviceProfilesIDs = append(deviceProfilesIDs, dpID)
	}

	applications := []storage.Application{
		{OrganizationID: organizations[0].ID, Name: "application-1", ServiceProfileID: serviceProfilesIDs[0]},
		{OrganizationID: organizations[1].ID, Name: "application-2", ServiceProfileID: serviceProfilesIDs[0]},
	}
	for i := range applications {
		if err := storage.CreateApplication(context.Background(), storage.DB(), &applications[i]); err != nil {
			t.Fatal(err)
		}
	}

	devices := []storage.Device{
		{DevEUI: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1}, Name: "test-1", ApplicationID: applications[0].ID, DeviceProfileID: deviceProfilesIDs[0]},
		{DevEUI: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2}, Name: "test-2", ApplicationID: applications[1].ID, DeviceProfileID: deviceProfilesIDs[1]},
	}
	for _, d := range devices {
		if err := storage.CreateDevice(context.Background(), storage.DB(), &d); err != nil {
			t.Fatal(err)
		}
	}

	fuotaDeployments := []storage.FUOTADeployment{
		{Name: "test-fuota", GroupType: storage.FUOTADeploymentGroupTypeC, DR: 5, Frequency: 868100000, Payload: []byte{1, 2, 3, 4}, FragSize: 20, MulticastTimeout: 1, UnicastTimeout: time.Second},
	}
	for i := range fuotaDeployments {
		if err := storage.CreateFUOTADeploymentForDevice(context.Background(), storage.DB(), &fuotaDeployments[i], devices[0].DevEUI); err != nil {
			t.Fatal(err)
		}
	}

	// cleanup once structs are in place
	users := []struct {
		ID       int64
		Username string
		IsActive bool
		IsAdmin  bool
	}{
		{ID: 11, Username: "user1", IsActive: true, IsAdmin: true},
		{ID: 12, Username: "user2", IsActive: true},
		{ID: 13, Username: "user3", IsActive: true},
		{ID: 14, Username: "user4", IsActive: true},
		{ID: 15, Username: "user5", IsActive: true},
		{ID: 16, Username: "user6", IsActive: true},
		{ID: 17, Username: "user7", IsActive: false},
		{ID: 18, Username: "user8", IsActive: false, IsAdmin: true},
		{ID: 19, Username: "user9", IsActive: true},
		{ID: 20, Username: "user10", IsActive: true},
		{ID: 21, Username: "user11", IsActive: false},
		{ID: 22, Username: "user12", IsActive: true},
	}
	for _, user := range users {
		_, err := storage.DB().Exec(`insert into "user" (id, created_at, updated_at, username, password_hash, session_ttl, is_active, is_admin) values ($1, now(), now(), $2, '', 0, $3, $4)`, user.ID, user.Username, user.IsActive, user.IsAdmin)
		if err != nil {
			t.Fatal(err)
		}
	}

	orgUsers := []struct {
		UserID         int64
		OrganizationID int64
		IsAdmin        bool
		IsDeviceAdmin  bool
		IsGatewayAdmin bool
	}{
		{UserID: users[8].ID, OrganizationID: organizations[0].ID, IsAdmin: false},
		{UserID: users[9].ID, OrganizationID: organizations[0].ID, IsAdmin: true},
		{UserID: users[10].ID, OrganizationID: organizations[0].ID, IsAdmin: false},
		{UserID: users[11].ID, OrganizationID: organizations[1].ID, IsAdmin: true},
	}
	for _, orgUser := range orgUsers {
		if err := storage.CreateOrganizationUser(context.Background(), storage.DB(), orgUser.OrganizationID, orgUser.UserID, orgUser.IsAdmin, orgUser.IsDeviceAdmin, orgUser.IsGatewayAdmin); err != nil {
			t.Fatal(err)
		}
	}

	gateways := []storage.Gateway{
		{MAC: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1}, Name: "gateway1", OrganizationID: organizations[0].ID, NetworkServerID: networkServers[0].ID},
		{MAC: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2}, Name: "gateway2", OrganizationID: organizations[1].ID, NetworkServerID: networkServers[0].ID},
	}
	for i := range gateways {
		if err := storage.CreateGateway(context.Background(), storage.DB(), &gateways[i]); err != nil {
			t.Fatal(err)
		}
	}

	Convey("Given a set of test users, applications and devices", t, func() {

		Convey("When testing ValidateUsersAccess (DisableAssignExistingUsers=false)", func() {
			DisableAssignExistingUsers = false
			tests := []validatorTest{
				{
					Name:       "global admin user can create and list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "inactive global admin user can not create or list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user8"},
					ExpectedOK: false,
				},
				{
					Name:       "organization admin user can create, and list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "normal user can not create or list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}
			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateUsersAccess (DisableAssignExistingUsers=true)", func() {
			DisableAssignExistingUsers = true
			tests := []validatorTest{
				{
					Name:       "global admin user can create and list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin user can create",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin user can not list",
					Validators: []ValidatorFunc{ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
			}
			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateUserAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user has access to read, update and delete",
					Validators: []ValidatorFunc{ValidateUserAccess(14, Read), ValidateUserAccess(14, Update), ValidateUserAccess(14, Delete)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "user itself has access to read",
					Validators: []ValidatorFunc{ValidateUserAccess(14, Read)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "user itself has no access to update or delete",
					Validators: []ValidatorFunc{ValidateUserAccess(14, Update), ValidateUserAccess(14, Delete)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
				{
					Name:       "other users are not able to read, update or delete",
					Validators: []ValidatorFunc{ValidateUserAccess(14, Read), ValidateUserAccess(14, Update), ValidateUserAccess(14, Delete)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateIsOrganizationAdmin", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users are",
					Validators: []ValidatorFunc{ValidateIsOrganizationAdmin(organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users are",
					Validators: []ValidatorFunc{ValidateIsOrganizationAdmin(organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "normal organization users are not",
					Validators: []ValidatorFunc{ValidateIsOrganizationAdmin(applications[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateOrganizationsAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create), ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "normal users users can list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can not create",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not create",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
				{
					Name:       "inactive global admin users can not create and list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create), ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user8"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateOrganizationAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read, update and delete",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Read, organizations[0].ID), ValidateOrganizationAccess(Update, organizations[0].ID), ValidateOrganizationAccess(Delete, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can read and update",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Read, organizations[0].ID), ValidateOrganizationAccess(Update, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can read",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Read, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin can not delete",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Delete, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
				{
					Name:       "organization users can not update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Update, organizations[0].ID), ValidateOrganizationAccess(Delete, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not read, update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Read, organizations[0].ID), ValidateOrganizationAccess(Update, organizations[0].ID), ValidateOrganizationAccess(Delete, organizations[0].ID)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateOrganizationUsersAccess (DisableAssignExistingUsers=false)", func() {
			DisableAssignExistingUsers = false
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID), ValidateOrganizationUsersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can create and list",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID), ValidateOrganizationUsersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can list",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can not create",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not create and list",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID), ValidateOrganizationUsersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateOrganizationUsersAccess (DisableAssignExistingUsers=true)", func() {
			DisableAssignExistingUsers = true
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID), ValidateOrganizationUsersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can not create",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateOrganizationUserAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read, update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationUserAccess(Read, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Update, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Delete, organizations[0].ID, users[8].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can read, update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationUserAccess(Read, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Update, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Delete, organizations[0].ID, users[8].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization user can read own user record",
					Validators: []ValidatorFunc{ValidateOrganizationUserAccess(Read, organizations[0].ID, users[8].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization user can not read other user record",
					Validators: []ValidatorFunc{ValidateOrganizationUserAccess(Read, organizations[0].ID, users[9].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "organization users can not update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationUserAccess(Update, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Delete, organizations[0].ID, users[8].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not read, update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationUserAccess(Read, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Update, organizations[0].ID, users[8].ID), ValidateOrganizationUserAccess(Delete, organizations[0].ID, users[8].ID)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("WHen testing ValidateGatewayProfileAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create, update, delete read and list",
					Validators: []ValidatorFunc{ValidateGatewayProfileAccess(Create), ValidateGatewayProfileAccess(Update), ValidateGatewayProfileAccess(Delete), ValidateGatewayProfileAccess(Read), ValidateGatewayProfileAccess(List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "normal users can read and list",
					Validators: []ValidatorFunc{ValidateGatewayProfileAccess(Read), ValidateGatewayProfileAccess(List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "normal users can not create, update and delete",
					Validators: []ValidatorFunc{ValidateGatewayProfileAccess(Create), ValidateGatewayProfileAccess(Update), ValidateGatewayProfileAccess(Delete)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateNetworkServersAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateNetworkServersAccess(Create, organizations[0].ID), ValidateNetworkServersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can list",
					Validators: []ValidatorFunc{ValidateNetworkServersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can not create",
					Validators: []ValidatorFunc{ValidateNetworkServersAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization users can not create or list",
					Validators: []ValidatorFunc{ValidateNetworkServersAccess(Create, organizations[0].ID), ValidateNetworkServersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateOrganizationNetworkServerAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read",
					Validators: []ValidatorFunc{ValidateOrganizationNetworkServerAccess(Read, organizations[0].ID, networkServers[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can read",
					Validators: []ValidatorFunc{ValidateOrganizationNetworkServerAccess(Read, organizations[0].ID, networkServers[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can not read when the network-server is not linked to the organization",
					Validators: []ValidatorFunc{ValidateOrganizationNetworkServerAccess(Read, organizations[0].ID, networkServers[1].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization users can not read",
					Validators: []ValidatorFunc{ValidateOrganizationNetworkServerAccess(Read, organizations[0].ID, networkServers[0].ID)},
					Claims:     Claims{Username: "user12"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateServiceProfilesAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(Create, organizations[0].ID), ValidateServiceProfilesAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can list",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can list",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "any user can list when organization id = 0",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(List, 0)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can not create",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
				{
					Name:       "organization users can not create",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization can not create or list",
					Validators: []ValidatorFunc{ValidateServiceProfilesAccess(Create, organizations[1].ID), ValidateServiceProfilesAccess(List, organizations[1].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateServiceProfileAccess", func() {
			id := serviceProfilesIDs[0]

			tests := []validatorTest{
				{
					Name:       "global admin users can read, update and delete",
					Validators: []ValidatorFunc{ValidateServiceProfileAccess(Read, id), ValidateServiceProfileAccess(Update, id), ValidateServiceProfileAccess(Delete, id)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can read",
					Validators: []ValidatorFunc{ValidateServiceProfileAccess(Read, id)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can read",
					Validators: []ValidatorFunc{ValidateServiceProfileAccess(Read, id)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can not update or delete",
					Validators: []ValidatorFunc{ValidateServiceProfileAccess(Update, id), ValidateServiceProfileAccess(Delete, id)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
				{
					Name:       "organization users can not update or delete",
					Validators: []ValidatorFunc{ValidateServiceProfileAccess(Update, id), ValidateServiceProfileAccess(Delete, id)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization users can not read, update or delete",
					Validators: []ValidatorFunc{ValidateServiceProfileAccess(Read, id), ValidateServiceProfileAccess(Update, id), ValidateServiceProfileAccess(Delete, id)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing MulticastGroupsAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateMulticastGroupsAccess(Create, organizations[0].ID), ValidateMulticastGroupsAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can create and list",
					Validators: []ValidatorFunc{ValidateMulticastGroupsAccess(Create, organizations[0].ID), ValidateMulticastGroupsAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can list",
					Validators: []ValidatorFunc{ValidateMulticastGroupsAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can not create",
					Validators: []ValidatorFunc{ValidateMulticastGroupsAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization users can not create or list",
					Validators: []ValidatorFunc{ValidateMulticastGroupsAccess(Create, organizations[0].ID), ValidateMulticastGroupsAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user12"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing MulticastGroupAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read, update and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupAccess(Read, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Update, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can read, update, and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupAccess(Read, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Update, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can read",
					Validators: []ValidatorFunc{ValidateMulticastGroupAccess(Read, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can not update and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupAccess(Update, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization users can not read, update and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupAccess(Read, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Update, multicastGroupsIDs[0]), ValidateMulticastGroupAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user12"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateMulticastGroupQueueAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user can create, read, list and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupQueueAccess(Create, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Read, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(List, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can create, read, list and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupQueueAccess(Create, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Read, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(List, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can create, read, list and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupQueueAccess(Create, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Read, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(List, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "non-organization users can not create, list and delete",
					Validators: []ValidatorFunc{ValidateMulticastGroupQueueAccess(Create, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(List, multicastGroupsIDs[0]), ValidateMulticastGroupQueueAccess(Delete, multicastGroupsIDs[0])},
					Claims:     Claims{Username: "user12"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateFUOTADeploymentAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user can read",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentAccess(Read, fuotaDeployments[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin can read",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentAccess(Read, fuotaDeployments[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization user can read",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentAccess(Read, fuotaDeployments[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "non-organization user can not read",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentAccess(Read, fuotaDeployments[0].ID)},
					Claims:     Claims{Username: "user12"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})

		Convey("When testing ValidateFUOTADeploymentsAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user can create",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentsAccess(Create, applications[0].ID, lorawan.EUI64{}), ValidateFUOTADeploymentsAccess(Create, 0, devices[0].DevEUI)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin user can create",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentsAccess(Create, applications[0].ID, lorawan.EUI64{}), ValidateFUOTADeploymentsAccess(Create, 0, devices[0].DevEUI)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization user can not create",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentsAccess(Create, applications[0].ID, lorawan.EUI64{}), ValidateFUOTADeploymentsAccess(Create, 0, devices[0].DevEUI)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "non-organization user can not create",
					Validators: []ValidatorFunc{ValidateFUOTADeploymentsAccess(Create, applications[0].ID, lorawan.EUI64{}), ValidateFUOTADeploymentsAccess(Create, 0, devices[0].DevEUI)},
					Claims:     Claims{Username: "user12"},
					ExpectedOK: false,
				},
			}

			runTests(tests, storage.DB())
		})
	})
}

func runTests(tests []validatorTest, db sqlx.Ext) {
	for i, test := range tests {
		Convey(fmt.Sprintf("testing: %s [%d]", test.Name, i), func() {
			for _, v := range test.Validators {
				ok, err := v(storage.DB(), &test.Claims)
				So(err, ShouldBeNil)
				So(ok, ShouldEqual, test.ExpectedOK)
			}
		})
	}
}
