package auth

import (
	"fmt"
	"testing"

	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	. "github.com/smartystreets/goconvey/convey"
)

type validatorTest struct {
	Name       string
	Claims     Claims
	Validators []ValidatorFunc
	ExpectedOK bool
}

func TestValidators(t *testing.T) {
	conf := test.GetConfig()

	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		t.Fatal(err)
	}
	test.MustResetDB(db)

	/*
		Users:
		1: global admin
		2: admin of application 1
		3: member of application 1
		4: no membership
		5: admin of application 2
		6: member of application 2
		7: member of application 2 (but is_active=false)
		8: global admin (but is_active=false)
		9: member of organization 1
		10: admin of organization 1
		11: member of organization 1 (but is_active=false)

		Organizations:
		1: organization 1
		2: organization 2

		Applications:
		1: application 1
		2: application 2

		Nodes:
		0101010101010101: application 1 node
		0202020202020202: application 2 node

		Gateways:
		0101010101010101: organization 1 gw
		0202020202020202: organization 2 gw
	*/
	organizations := []storage.Organization{
		{Name: "organization-1"},
		{Name: "organization-2"},
	}
	for i := range organizations {
		if err := storage.CreateOrganization(db, &organizations[i]); err != nil {
			t.Fatal(err)
		}
	}

	applications := []storage.Application{
		{OrganizationID: organizations[0].ID, Name: "application-1"},
		{OrganizationID: organizations[1].ID, Name: "application-2"},
	}
	for i := range applications {
		if err := storage.CreateApplication(db, &applications[i]); err != nil {
			t.Fatal(err)
		}
	}

	nodes := []storage.Node{
		{DevEUI: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1}, Name: "test-1", ApplicationID: applications[0].ID},
		{DevEUI: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2}, Name: "test-2", ApplicationID: applications[1].ID},
	}
	for _, node := range nodes {
		if err := storage.CreateNode(db, node); err != nil {
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
	}
	for _, user := range users {
		_, err = db.Exec(`insert into "user" (id, created_at, updated_at, username, password_hash, session_ttl, is_active, is_admin) values ($1, now(), now(), $2, '', 0, $3, $4)`, user.ID, user.Username, user.IsActive, user.IsAdmin)
		if err != nil {
			t.Fatal(err)
		}
	}

	appUsers := []struct {
		UserID        int64
		ApplicationID int64
		IsAdmin       bool
	}{
		{UserID: 12, ApplicationID: applications[0].ID, IsAdmin: true},
		{UserID: 13, ApplicationID: applications[0].ID, IsAdmin: false},
		{UserID: 15, ApplicationID: applications[1].ID, IsAdmin: true},
		{UserID: 16, ApplicationID: applications[1].ID, IsAdmin: false},
		{UserID: 17, ApplicationID: applications[1].ID, IsAdmin: false},
	}
	for _, appUser := range appUsers {
		_, err = db.Exec("insert into application_user (created_at, updated_at, user_id, application_id, is_admin) values (now(), now(), $1, $2, $3)", appUser.UserID, appUser.ApplicationID, appUser.IsAdmin)
		if err != nil {
			t.Fatal(err)
		}
	}

	orgUsers := []struct {
		UserID         int64
		OrganizationID int64
		IsAdmin        bool
	}{
		{UserID: users[8].ID, OrganizationID: organizations[0].ID, IsAdmin: false},
		{UserID: users[9].ID, OrganizationID: organizations[0].ID, IsAdmin: true},
		{UserID: users[10].ID, OrganizationID: organizations[0].ID, IsAdmin: false},
	}
	for _, orgUser := range orgUsers {
		if err := storage.CreateOrganizationUser(db, orgUser.OrganizationID, orgUser.UserID, orgUser.IsAdmin); err != nil {
			t.Fatal(err)
		}
	}

	gateways := []storage.Gateway{
		{MAC: lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1}, Name: "gateway1", OrganizationID: organizations[0].ID},
		{MAC: lorawan.EUI64{2, 2, 2, 2, 2, 2, 2, 2}, Name: "gateway2", OrganizationID: organizations[1].ID},
	}
	for i := range gateways {
		if err := storage.CreateGateway(db, &gateways[i]); err != nil {
			t.Fatal(err)
		}
	}

	Convey("Given a set of test users, applications and nodes", t, func() {

		Convey("When testing ValidateUsersAccess", func() {
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
					Name:       "application admin user can create, and list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: true,
				},
				{
					Name:       "normal user can not create or list",
					Validators: []ValidatorFunc{ValidateUsersAccess(Create), ValidateUsersAccess(List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}
			runTests(tests, db)
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

			runTests(tests, db)
		})

		Convey("When testing ValidateApplicationsAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user has access to create and list",
					Validators: []ValidatorFunc{ValidateApplicationsAccess(Create), ValidateApplicationsAccess(List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "other users don't have access to create",
					Validators: []ValidatorFunc{ValidateApplicationsAccess(Create)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: false,
				},
				{
					Name:       "normal user has access to list",
					Validators: []ValidatorFunc{ValidateApplicationsAccess(List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "inactive users do not have access to create or list",
					Validators: []ValidatorFunc{ValidateApplicationsAccess(Create), ValidateApplicationsAccess(List)},
					Claims:     Claims{Username: "user7"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("WHen testing ValidateApplicationAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read, update and delete",
					Validators: []ValidatorFunc{ValidateApplicationAccess(1, Read), ValidateApplicationAccess(1, Update), ValidateApplicationAccess(1, Delete)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin users can read and update",
					Validators: []ValidatorFunc{ValidateApplicationAccess(1, Read), ValidateApplicationAccess(1, Update)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: true,
				},
				{
					Name:       "application users can read",
					Validators: []ValidatorFunc{ValidateApplicationAccess(1, Read)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin users can not delete",
					Validators: []ValidatorFunc{ValidateApplicationAccess(1, Delete)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: false,
				},
				{
					Name:       "application users can not update or delete",
					Validators: []ValidatorFunc{ValidateApplicationAccess(1, Update), ValidateApplicationAccess(1, Delete)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
				{
					Name:       "non-application users can not read, update or delete",
					Validators: []ValidatorFunc{ValidateApplicationAccess(1, Read), ValidateApplicationAccess(1, Update), ValidateApplicationAccess(1, Delete)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateApplicationMembersAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user has access to create and list",
					Validators: []ValidatorFunc{ValidateApplicationMembersAccess(1, Create), ValidateApplicationMembersAccess(1, List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin user has access to create and list",
					Validators: []ValidatorFunc{ValidateApplicationMembersAccess(1, Create), ValidateApplicationMembersAccess(1, List)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin user of different application has no access to create or list",
					Validators: []ValidatorFunc{ValidateApplicationMembersAccess(1, Create), ValidateApplicationMembersAccess(1, List)},
					Claims:     Claims{Username: "user6"},
					ExpectedOK: false,
				},
				{
					Name:       "application users are able to list",
					Validators: []ValidatorFunc{ValidateApplicationMembersAccess(1, List)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: true,
				},
				{
					Name:       "application users are not able to create",
					Validators: []ValidatorFunc{ValidateApplicationMembersAccess(1, Create)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateApplicationMemberAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user has access to read, update and delete",
					Validators: []ValidatorFunc{ValidateApplicationMemberAccess(1, 13, Read), ValidateApplicationMemberAccess(1, 13, Update), ValidateApplicationMemberAccess(1, 13, Delete)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin user has access to read, update and delete",
					Validators: []ValidatorFunc{ValidateApplicationMemberAccess(1, 13, Read), ValidateApplicationMemberAccess(1, 13, Update), ValidateApplicationMemberAccess(1, 13, Delete)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin user of different application has no access to read, update or delete",
					Validators: []ValidatorFunc{ValidateApplicationMemberAccess(1, 13, Read), ValidateApplicationMemberAccess(1, 13, Update), ValidateApplicationMemberAccess(1, 13, Delete)},
					Claims:     Claims{Username: "user6"},
					ExpectedOK: false,
				},
				{
					Name:       "application users are not able to update or delete",
					Validators: []ValidatorFunc{ValidateApplicationMemberAccess(1, 13, Update), ValidateApplicationMemberAccess(1, 13, Delete)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
				{
					Name:       "application user is able to read its own record",
					Validators: []ValidatorFunc{ValidateApplicationMemberAccess(1, 13, Read)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: true,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateNodesAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin user has access to create and list",
					Validators: []ValidatorFunc{ValidateNodesAccess(1, Create), ValidateNodesAccess(1, List)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin users can create and list",
					Validators: []ValidatorFunc{ValidateNodesAccess(1, Create), ValidateNodesAccess(1, List)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: true,
				},
				{
					Name:       "application users can list",
					Validators: []ValidatorFunc{ValidateNodesAccess(1, List)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: true,
				},
				{
					Name:       "application users can not create",
					Validators: []ValidatorFunc{ValidateNodesAccess(1, Create)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
				{
					Name:       "other users can not create or list",
					Validators: []ValidatorFunc{ValidateNodesAccess(1, Create), ValidateNodesAccess(1, List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateNodeAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read, update and delete",
					Validators: []ValidatorFunc{ValidateNodeAccess(nodes[0].DevEUI, Read), ValidateNodeAccess(nodes[0].DevEUI, Update), ValidateNodeAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "application admin users can read, update and delete",
					Validators: []ValidatorFunc{ValidateNodeAccess(nodes[0].DevEUI, Read), ValidateNodeAccess(nodes[0].DevEUI, Update), ValidateNodeAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user2"},
					ExpectedOK: true,
				},
				{
					Name:       "application users can read",
					Validators: []ValidatorFunc{ValidateNodeAccess(nodes[0].DevEUI, Read)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: true,
				},
				{

					Name:       "application users (non-admin) can not update or delete",
					Validators: []ValidatorFunc{ValidateNodeAccess(nodes[0].DevEUI, Update), ValidateNodeAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
				{
					Name:       "other users can not read, update and delete",
					Validators: []ValidatorFunc{ValidateNodeAccess(nodes[0].DevEUI, Read), ValidateNodeAccess(nodes[0].DevEUI, Update), ValidateNodeAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateNodeQueueAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can read, list, update and delete",
					Validators: []ValidatorFunc{ValidateNodeQueueAccess(nodes[0].DevEUI, Create), ValidateNodeQueueAccess(nodes[0].DevEUI, Read), ValidateNodeQueueAccess(nodes[0].DevEUI, List), ValidateNodeQueueAccess(nodes[0].DevEUI, Update), ValidateNodeQueueAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "application users can read, list, update and delete",
					Validators: []ValidatorFunc{ValidateNodeQueueAccess(nodes[0].DevEUI, Read), ValidateNodeQueueAccess(nodes[0].DevEUI, List), ValidateNodeQueueAccess(nodes[0].DevEUI, Update), ValidateNodeQueueAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: true,
				},
				{
					Name:       "other users can not read, list, update and delete",
					Validators: []ValidatorFunc{ValidateNodeQueueAccess(nodes[0].DevEUI, Create), ValidateNodeQueueAccess(nodes[0].DevEUI, Read), ValidateNodeQueueAccess(nodes[0].DevEUI, List), ValidateNodeQueueAccess(nodes[0].DevEUI, Update), ValidateNodeQueueAccess(nodes[0].DevEUI, Delete)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateChannelListAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create, read, list, update and delete",
					Validators: []ValidatorFunc{ValidateChannelListAccess(Create), ValidateChannelListAccess(Read), ValidateChannelListAccess(List), ValidateChannelListAccess(Update), ValidateChannelListAccess(Delete)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "users are able to read and list",
					Validators: []ValidatorFunc{ValidateChannelListAccess(Read), ValidateChannelListAccess(List)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "users are not able to create update or delete",
					Validators: []ValidatorFunc{ValidateChannelListAccess(Create), ValidateChannelListAccess(Update), ValidateChannelListAccess(Delete)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateGatewaysAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create and list",
					Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, organizations[0].ID), ValidateGatewaysAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can create and list",
					Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, organizations[0].ID), ValidateGatewaysAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "normal user can list",
					Validators: []ValidatorFunc{ValidateGatewaysAccess(List, 0)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: true,
				},
				{
					Name:       "organization user can not create",
					Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "normal user can not create",
					Validators: []ValidatorFunc{ValidateGatewaysAccess(Create, organizations[0].ID)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
				{
					Name:       "inactive user can not list",
					Validators: []ValidatorFunc{ValidateGatewaysAccess(List, 0)},
					Claims:     Claims{Username: "user11"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateGatewayAccess", func() {
			tests := []validatorTest{
				{
					Name:       "global admin users can create, update and delete",
					Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
					Claims:     Claims{Username: "user1"},
					ExpectedOK: true,
				},
				{
					Name:       "organization admin users can create, update and delete",
					Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can read",
					Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: true,
				},
				{
					Name:       "organization users can not update or delete",
					Validators: []ValidatorFunc{ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not read, update or delete",
					Validators: []ValidatorFunc{ValidateGatewayAccess(Read, gateways[0].MAC), ValidateGatewayAccess(Update, gateways[0].MAC), ValidateGatewayAccess(Delete, gateways[0].MAC)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
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
					Name:       "organization admin users can not create and list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create), ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user10"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not create and list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create), ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user9"},
					ExpectedOK: false,
				},
				{
					Name:       "inactive global admin users can not create and list",
					Validators: []ValidatorFunc{ValidateOrganizationsAccess(Create), ValidateOrganizationsAccess(List)},
					Claims:     Claims{Username: "user8"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
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
					Name:       "application users within an organization can read",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Read, organizations[0].ID)},
					Claims:     Claims{Username: "user3"},
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
					Name:       "application users within an an organization can not update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Update, organizations[0].ID), ValidateOrganizationAccess(Delete, organizations[0].ID)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
				{
					Name:       "normal users can not read, update or delete",
					Validators: []ValidatorFunc{ValidateOrganizationAccess(Read, organizations[0].ID), ValidateOrganizationAccess(Update, organizations[0].ID), ValidateOrganizationAccess(Delete, organizations[0].ID)},
					Claims:     Claims{Username: "user4"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
		})

		Convey("When testing ValidateOrganizationUsersAccess", func() {
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
				{
					Name:       "application users within an organization can not create and list",
					Validators: []ValidatorFunc{ValidateOrganizationUsersAccess(Create, organizations[0].ID), ValidateOrganizationUsersAccess(List, organizations[0].ID)},
					Claims:     Claims{Username: "user3"},
					ExpectedOK: false,
				},
			}

			runTests(tests, db)
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

			runTests(tests, db)
		})
	})
}

func runTests(tests []validatorTest, db *sqlx.DB) {
	for i, test := range tests {
		Convey(fmt.Sprintf("testing: %s [%d]", test.Name, i), func() {
			for _, v := range test.Validators {
				ok, err := v(db, &test.Claims)
				So(err, ShouldBeNil)
				So(ok, ShouldEqual, test.ExpectedOK)
			}
		})
	}
}
