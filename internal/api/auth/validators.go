package auth

import (
	"strings"

	"github.com/gofrs/uuid"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Flag defines the authorization flag.
type Flag int

// DisableAssignExistingUsers controls if existing users can be assigned
// to an organization or application. When set to false (default), organization
// admin users are able to list all users, which might depending on the
// context of the setup be a privacy issue.
var DisableAssignExistingUsers = false

// Authorization flags.
const (
	Create Flag = iota
	Read
	Update
	Delete
	List
	UpdateProfile
)

const userQuery = `
	select count(*)
	from "user" u
	left join organization_user ou
		on u.id = ou.user_id
	left join organization o
		on o.id = ou.organization_id
	left join gateway g
		on o.id = g.organization_id
	left join application a
		on a.organization_id = o.id
	left join service_profile sp
		on sp.organization_id = o.id
	left join device_profile dp
		on dp.organization_id = o.id
	left join network_server ns
		on ns.id = sp.network_server_id or ns.id = dp.network_server_id
	left join device d
		on a.id = d.application_id
	left join multicast_group mg
		on sp.service_profile_id = mg.service_profile_id
`

// ValidateActiveUser validates if the user in the JWT claim is active.
func ValidateActiveUser() ValidatorFunc {
	where := [][]string{
		{"u.username = $1", "u.is_active = true"},
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateUsersAccess validates if the client has access to the global users
// resource.
func ValidateUsersAccess(flag Flag) ValidatorFunc {
	var where [][]string

	switch flag {
	case Create:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true"},
		}
	case List:
		if DisableAssignExistingUsers {
			// global admin users
			where = [][]string{
				{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			}
		} else {
			// global admin
			// organization admin
			where = [][]string{
				{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
				{"u.username = $1", "u.is_active = true", "ou.is_admin = true"},
			}
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateUserAccess validates if the client has access to the given user
// resource.
func ValidateUserAccess(userID int64, flag Flag) ValidatorFunc {
	var where [][]string

	switch flag {
	case Read:
		// global admin
		// user itself
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "u.id = $2"},
		}
	case Update, Delete:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	case UpdateProfile:
		// global admin
		// user itself
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "u.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, userID)
	}
}

// ValidateIsApplicationAdmin validates if the client has access to
// administrate the given application.
func ValidateIsApplicationAdmin(applicationID int64) ValidatorFunc {
	// global admin
	// organization admin
	where := [][]string{
		{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
		{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateApplicationsAccess validates if the client has access to the
// global applications resource.
func ValidateApplicationsAccess(flag Flag, organizationID int64) ValidatorFunc {
	var where [][]string

	switch flag {
	case Create:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}
	case List:
		// global admin
		// organization user (when organization id is given)
		// any active user (api will filter on user)
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "$2 > 0", "o.id = $2 or a.organization_id = $2"},
			{"u.username = $1", "u.is_active = true", "$2 = 0"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID)
	}
}

// ValidateApplicationAccess validates if the client has access to the given
// application.
func ValidateApplicationAccess(applicationID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "a.id = $2"},
		}
	case Update:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
		}
	case Delete:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateApplicationUsersAccess validates if the client has access to the
// given application members.
func ValidateApplicationUsersAccess(applicationID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		if DisableAssignExistingUsers {
			// global admin
			where = [][]string{
				{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
			}
		} else {
			// global admin
			// organization admin
			where = [][]string{
				{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
				{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
			}
		}
	case List:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateApplicationUserAccess validates if the client has access to the
// given application member.
func ValidateApplicationUserAccess(applicationID, userID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization admin
		// user itself
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
			{"u.username = $1", "u.is_active = true", "a.id = $2", "ou.user_id = $3"},
		}
	case Update:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
		}
	case Delete:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID, userID)
	}
}

// ValidateNodesAccess validates if the client has access to the global nodes
// resource.
func ValidateNodesAccess(applicationID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
		}
	case List:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateNodeAccess validates if the client has access to the given node.
func ValidateNodeAccess(devEUI lorawan.EUI64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "d.dev_eui = $2"},
		}
	case Update:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "d.dev_eui = $2"},
		}
	case Delete:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "d.dev_eui = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, devEUI[:])
	}
}

// ValidateDeviceQueueAccess validates if the client has access to the queue
// of the given node.
func ValidateDeviceQueueAccess(devEUI lorawan.EUI64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create, List, Delete:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "d.dev_eui = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, devEUI[:])
	}
}

// ValidateGatewaysAccess validates if the client has access to the gateways.
func ValidateGatewaysAccess(flag Flag, organizationID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true", "o.can_have_gateways = true"},
		}
	case List:
		// global admin
		// organization user
		// any active user (result filtered on user)
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "$2 > 0", "o.id = $2"},
			{"u.username = $1", "u.is_active = true", "$2 = 0"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID)
	}
}

// ValidateGatewayAccess validates if the client has access to the given gateway.
func ValidateGatewayAccess(flag Flag, mac lorawan.EUI64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "g.mac = $2"},
		}
	case Update, Delete:
		where = [][]string{
			// global admin
			// organization admin
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "g.mac = $2", "ou.is_admin = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, mac[:])
	}
}

// ValidateIsOrganizationAdmin validates if the client has access to
// administrate the given organization.
func ValidateIsOrganizationAdmin(organizationID int64) ValidatorFunc {
	// global admin
	// organization admin
	where := [][]string{
		{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
		{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "o.id = $2"},
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID)
	}
}

// ValidateOrganizationsAccess validates if the client has access to the
// organizations.
func ValidateOrganizationsAccess(flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
		}
	case List:
		// any active user (results are filtered by the api)
		where = [][]string{
			{"u.username = $1", "u.is_active = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateOrganizationAccess validates if the client has access to the
// given organization.
func ValidateOrganizationAccess(flag Flag, id int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2"},
			{"u.username = $1", "u.is_active = true", "a.organization_id = $2"},
		}
	case Update:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}
	case Delete:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, id)
	}
}

// ValidateOrganizationUsersAccess validates if the client has access to
// the organization users.
func ValidateOrganizationUsersAccess(flag Flag, id int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		if DisableAssignExistingUsers {
			// global admin
			where = [][]string{
				{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
			}
		} else {
			// global admin
			// organization admin
			where = [][]string{
				{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
				{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
			}
		}
	case List:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, id)
	}
}

// ValidateOrganizationUserAccess validates if the client has access to the
// given user of the given organization.
func ValidateOrganizationUserAccess(flag Flag, organizationID, userID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization admin
		// user itself
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.user_id = $3", "ou.user_id = u.id"},
		}
	case Update:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}
	case Delete:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID, userID)
	}
}

// ValidateGatewayProfileAccess validates if the client has access
// to the gateway-profiles.
func ValidateGatewayProfileAccess(flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create, Update, Delete:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
		}
	case Read, List:
		// any active user
		where = [][]string{
			{"u.username = $1", "u.is_active = true"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateNetworkServersAccess validates if the client has access to the
// network-servers.
func ValidateNetworkServersAccess(flag Flag, organizationID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	case List:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID)
	}
}

// ValidateNetworkServerAccess validates if the client has access to the
// given network-server.
func ValidateNetworkServerAccess(flag Flag, id int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// org. admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "ns.id = $2"},
		}
	case Update, Delete:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, id)
	}
}

// ValidateOrganizationNetworkServerAccess validates if the given client has
// access to the given organization id / network server id combination.
func ValidateOrganizationNetworkServerAccess(flag Flag, organizationID, networkServerID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ns.id = $3"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID, networkServerID)
	}
}

// ValidateServiceProfilesAccess validates if the client has access to the
// service-profiles.
func ValidateServiceProfilesAccess(flag Flag, organizationID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	case List:
		// global admin
		// organization user (when organization id is given)
		// any active user (filtered by user)
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "$2 > 0", "o.id = $2"},
			{"u.username = $1", "u.is_active = true", "$2 = 0"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID)
	}
}

// ValidateServiceProfileAccess validates if the client has access to the
// given service-profile.
func ValidateServiceProfileAccess(flag Flag, id uuid.UUID) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization users to which the service-profile is linked
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "sp.service_profile_id = $2"},
		}
	case Update, Delete:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, id)
	}
}

// ValidateDeviceProfilesAccess validates if the client has access to the
// device-profiles.
func ValidateDeviceProfilesAccess(flag Flag, organizationID, applicationID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true", "$3 = 0"},
		}
	case List:
		// global admin
		// organization user (when organization id is given)
		// user linked to a given application (when application id is given)
		// any active user (filtered by user)
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "$3 = 0", "$2 > 0", "o.id = $2"},
			{"u.username = $1", "u.is_active = true", "$2 = 0", "$3 > 0", "a.id = $3"},
			{"u.username = $1", "u.is_active = true", "$2 = 0", "$3 = 0"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID, applicationID)
	}
}

// ValidateDeviceProfileAccess validates if the client has access to the
// given device-profile.
func ValidateDeviceProfileAccess(flag Flag, id uuid.UUID) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// gloabal admin
		// organization users
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "dp.device_profile_id = $2"},
		}
	case Update, Delete:
		// global admin
		// organization admin users
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin=true", "dp.device_profile_id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, id)
	}
}

// ValidateMulticastGroupsAccess validates if the client has access to the
// multicast-groups.
func ValidateMulticastGroupsAccess(flag Flag, organizationID int64) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}

	case List:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "o.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, organizationID)
	}
}

// ValidateMulticastGroupAccess validates if the client has access to the given
// multicast-group.
func ValidateMulticastGroupAccess(flag Flag, multicastGroupID uuid.UUID) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization users
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "mg.id = $2"},
		}
	case Update, Delete:
		// global admin
		// organization admin users
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "ou.is_admin = true", "mg.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, multicastGroupID)
	}
}

// ValidateMulticastGroupQueueAccess validates if the client has access to
// the given multicast-group queue.
func ValidateMulticastGroupQueueAccess(flag Flag, multicastGroupID uuid.UUID) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create, Read, List, Delete:
		// global admin
		// organization user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "mg.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, multicastGroupID)
	}
}

func executeQuery(db sqlx.Queryer, query string, where [][]string, args ...interface{}) (bool, error) {
	var ors []string
	for _, ands := range where {
		ors = append(ors, "(("+strings.Join(ands, ") and (")+"))")
	}
	whereStr := strings.Join(ors, " or ")
	query = query + " where " + whereStr

	var count int64
	if err := sqlx.Get(db, &count, query, args...); err != nil {
		return false, errors.Wrap(err, "select error")
	}
	return count > 0, nil
}
