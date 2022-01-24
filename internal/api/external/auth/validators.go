package auth

import (
	"strings"

	"github.com/gofrs/uuid"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// API key subjects.
const (
	SubjectUser   = "user"
	SubjectAPIKey = "api_key"
)

// Flag defines the authorization flag.
type Flag int

// Authorization flags.
const (
	Create Flag = iota
	Read
	Update
	Delete
	List
	UpdateProfile
	ADRAlgorithms
)

// ValidateActiveUser validates if the user in the JWT claim is active.
func ValidateActiveUser() ValidatorFunc {
	query := `
		select
			1
		from
			"user" u
	`

	where := [][]string{
		{"(u.email = $1 or u.id = $2)", "u.is_active = true"},
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, query, where, claims.Username, claims.UserID)
		case SubjectAPIKey:
			return false, nil
		default:
			return false, nil
		}
	}
}

// ValidateUsersAccess validates if the client has access to the global users
// resource.
func ValidateUsersAccess(flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere [][]string
	var apiKeyWhere [][]string

	switch flag {
	case Create:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $2)", "u.is_active = true", "u.is_admin = true"},
		}

		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	case List:
		// global admin users
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $2)", "u.is_active = true", "u.is_admin = true"},
		}

		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID)
		default:
			return false, nil
		}
	}
}

// ValidateUserAccess validates if the client has access to the given user
// resource.
func ValidateUserAccess(userID int64, flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere [][]string
	var apiKeyWhere [][]string

	switch flag {
	case Read:
		// global admin
		// user itself
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.id = $2"},
		}

		// admin token
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	case Update, Delete:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}

		// admin token
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	case UpdateProfile:
		// global admin
		// user itself
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.id = $2"},
		}

		// admin token
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, userID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID)
		default:
			return false, nil
		}
	}
}

// ValidateApplicationsAccess validates if the client has access to the
// global applications resource.
func ValidateApplicationsAccess(flag Flag, organizationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join application a
			on a.organization_id = o.id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.application_id = a.id
		left join organization o
			on ak.organization_id = o.id
	`

	var userWhere [][]string
	var apiKeyWhere [][]string

	switch flag {
	case Create:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2", "ou.is_device_admin = true"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "o.id = $2"},
		}
	case List:
		// global admin
		// organization user (when organization id is given)
		// any active user (api will filter on user)
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "$2 > 0", "o.id = $2 or a.organization_id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "$2 = 0"},
		}

		// admin api key
		// organization api key (api will do filtering)
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin"},
			{"ak.id = $1", "o.id = $2"},
		}

	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID)
		default:
			return false, nil
		}
	}
}

// ValidateApplicationAccess validates if the client has access to the given
// application.
func ValidateApplicationAccess(applicationID int64, flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join application a
			on a.organization_id = o.id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.application_id = a.id or ak.organization_id = a.organization_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "a.id = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"}, // application is joined on both a.id and a.organization_id
		}
	case Update:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_device_admin = true", "a.id = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"}, // application is joined on both a.id and a.organization_id
		}
	case Delete:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_device_admin = true", "a.id = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"}, // application is joined on both a.id and a.organization_id
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, applicationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, applicationID)
		default:
			return false, nil
		}
	}
}

// ValidateNodesAccess validates if the client has access to the global nodes
// resource.
func ValidateNodesAccess(applicationID int64, flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join application a
			on a.organization_id = o.id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.application_id = a.id or ak.organization_id = a.organization_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "a.id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_device_admin = true", "a.id = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"}, // application is joined on a.id and a.organization_id
		}
	case List:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "a.id = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"}, // application is joined on a.id and a.organization_id
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, applicationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, applicationID)
		default:
			return false, nil
		}
	}
}

// ValidateNodeAccess validates if the client has access to the given node.
func ValidateNodeAccess(devEUI lorawan.EUI64, flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join application a
			on a.organization_id = o.id
		left join device d
			on a.id = d.application_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.application_id = a.id or ak.organization_id = a.organization_id
		left join device d
			on a.id = d.application_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "d.dev_eui = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "d.dev_eui = $2"}, // application is joined on a.id and a.organization_id
		}

	case Update:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "d.dev_eui = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_device_admin = true", "d.dev_eui = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "d.dev_eui = $2"}, // application is joined on a.id and a.organization_id
		}
	case Delete:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "d.dev_eui = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_device_admin = true", "d.dev_eui = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "d.dev_eui = $2"}, // application is joined on a.id and a.organization_id
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, devEUI[:], claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, devEUI[:])
		default:
			return false, nil
		}
	}
}

// ValidateDeviceQueueAccess validates if the client has access to the queue
// of the given node.
func ValidateDeviceQueueAccess(devEUI lorawan.EUI64, flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join application a
			on a.organization_id = ou.organization_id
		left join device d
			on a.id = d.application_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.application_id = a.id or ak.organization_id = a.organization_id
		left join device d
			on a.id = d.application_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create, List, Delete:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "d.dev_eui = $2"},
		}

		// admin api key
		// organization api key
		// application api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "d.dev_eui = $2"}, // application is joined on a.id and a.organization_id
		}

	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, devEUI[:], claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, devEUI[:])
		default:
			return false, nil
		}
	}
}

// ValidateGatewaysAccess validates if the client has access to the gateways.
func ValidateGatewaysAccess(flag Flag, organizationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join organization o
			on ak.organization_id = o.id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		// gateway admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2", "ou.is_admin = true", "o.can_have_gateways = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2", "ou.is_gateway_admin = true", "o.can_have_gateways = true"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "o.id = $2", "o.can_have_gateways = true"},
		}
	case List:
		// global admin
		// organization user
		// any active user (result filtered on user)
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "$2 > 0", "o.id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "$2 = 0"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "o.id = $2"},
		}

	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID)
		default:
			return false, nil
		}
	}
}

// ValidateGatewayAccess validates if the client has access to the given gateway.
func ValidateGatewayAccess(flag Flag, mac lorawan.EUI64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join gateway g
			on o.id = g.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join gateway g
			on ak.organization_id = g.organization_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "g.mac = $2"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "g.mac = $2"},
		}
	case Update, Delete:
		userWhere = [][]string{
			// global admin
			// organization admin
			// organization gateway admin
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "g.mac = $2", "ou.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "g.mac = $2", "ou.is_gateway_admin = true"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "g.mac = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, mac[:], claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, mac[:])
		default:
			return false, nil
		}
	}
}

// ValidateIsOrganizationAdmin validates if the client has access to
// administrate the given organization.
func ValidateIsOrganizationAdmin(organizationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join organization o
			on ak.organization_id = o.id
	`

	// global admin
	// organization admin
	userWhere := [][]string{
		{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
		{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "o.id = $2"},
	}

	// admin api key
	// organization api key
	apiKeyWhere := [][]string{
		{"ak.id = $1", "ak.is_admin = true"},
		{"ak.id = $1", "o.id = $2"},
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID)
		default:
			return false, nil
		}
	}
}

// ValidateOrganizationsAccess validates if the client has access to the
// organizations.
func ValidateOrganizationsAccess(flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $2)", "u.is_active = true", "u.is_admin = true"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	case List:
		// any active user (results are filtered by the api)
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $2)", "u.is_active = true"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID)
		default:
			return false, nil
		}
	}
}

// ValidateOrganizationAccess validates if the client has access to the
// given organization.
func ValidateOrganizationAccess(flag Flag, id int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	case Update:
		// global admin
		// organization admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true", "$2 = $2"},
		}
	case Delete:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true", "$2 = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, id, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, id)
		default:
			return false, nil
		}
	}
}

// ValidateOrganizationUsersAccess validates if the client has access to
// the organization users.
func ValidateOrganizationUsersAccess(flag Flag, id int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	case List:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, id, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, id)
		default:
			return false, nil
		}
	}
}

// ValidateOrganizationUserAccess validates if the client has access to the
// given user of the given organization.
func ValidateOrganizationUserAccess(flag Flag, organizationID, userID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization admin
		// user itself
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ou.user_id = $3", "ou.user_id = u.id"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	case Update:
		// global admin
		// organization admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	case Delete:
		// global admin
		// organization admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ou.is_admin = true"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, userID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID)
		default:
			return false, nil
		}
	}
}

// ValidateGatewayProfileAccess validates if the client has access
// to the gateway-profiles.
func ValidateGatewayProfileAccess(flag Flag) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create, Update, Delete:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $2)", "u.is_active = true", "u.is_admin = true"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
		}
	case Read, List:
		// any active user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $2)", "u.is_active = true"},
		}

		// any api key
		apiKeyWhere = [][]string{
			{"ak.id = $1"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID)
		default:
			return false, nil
		}
	}
}

// ValidateNetworkServersAccess validates if the client has access to the
// network-servers.
func ValidateNetworkServersAccess(flag Flag, organizationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true", "$2 = $2"},
		}
	case List:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "o.id = $2"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID)
		default:
			return false, nil
		}
	}
}

// ValidateNetworkServerAccess validates if the client has access to the
// given network-server.
func ValidateNetworkServerAccess(flag Flag, id int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join service_profile sp
			on sp.organization_id = o.id
		left join network_server ns
			on ns.id = sp.network_server_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join service_profile sp
			on ak.organization_id = sp.organization_id
		left join network_server ns
			on sp.network_server_id = ns.id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization admin
		// organization gateway admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "ns.id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_gateway_admin = true", "ns.id = $2"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ns.id = $2"},
		}
	case Update, Delete:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true", "$2 = $2"},
		}
	case ADRAlgorithms:
		// global admin
		// active user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ns.id = $2"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ns.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, id, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, id)
		default:
			return false, nil
		}
	}
}

// ValidateOrganizationNetworkServerAccess validates if the given client has
// access to the given organization id / network server id combination.
func ValidateOrganizationNetworkServerAccess(flag Flag, organizationID, networkServerID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join service_profile sp
			on sp.organization_id = o.id
		left join device_profile dp
			on dp.organization_id = o.id
		left join network_server ns
			on ns.id = sp.network_server_id or ns.id = dp.network_server_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join service_profile sp
			on ak.organization_id = sp.organization_id
		left join network_server ns
			on sp.network_server_id = ns.id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ns.id = $3"},
		}

		// admin api key
		// organization api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2", "ns.id = $3"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, networkServerID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID, networkServerID)
		default:
			return false, nil
		}
	}
}

// ValidateServiceProfilesAccess validates if the client has access to the
// service-profiles.
func ValidateServiceProfilesAccess(flag Flag, organizationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true", "$2 = $2"},
		}
	case List:
		// global admin
		// organization user (when organization id is given)
		// any active user (filtered by user)
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "$2 > 0", "o.id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "$2 = 0"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID)
		default:
			return false, nil
		}
	}
}

// ValidateServiceProfileAccess validates if the client has access to the
// given service-profile.
func ValidateServiceProfileAccess(flag Flag, id uuid.UUID) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join service_profile sp
			on sp.organization_id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join service_profile sp
			on ak.organization_id = sp.organization_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization users to which the service-profile is linked
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "sp.service_profile_id = $2"},
		}

		// admin api key
		// org api key to which the service-profile is linked
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "sp.service_profile_id = $2"},
		}
	case Update, Delete:
		// global admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}

		// admin api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true", "$2 = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, id, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, id)
		default:
			return false, nil
		}
	}
}

// ValidateDeviceProfilesAccess validates if the client has access to the
// device-profiles.
func ValidateDeviceProfilesAccess(flag Flag, organizationID, applicationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join application a
			on a.organization_id = o.id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.organization_id = a.organization_id or ak.application_id = a.id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		// organization device admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ou.is_admin = true", "$3 = 0"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "o.id = $2", "ou.is_device_admin = true", "$3 = 0"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2", "$3 = $3"},
		}
	case List:
		// global admin
		// organization user (when organization id is given)
		// user linked to a given application (when application id is given)
		// any active user (filtered by user)
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "$3 = 0", "$2 > 0", "o.id = $2"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "$2 = 0", "$3 > 0", "a.id = $3"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "$2 = 0", "$3 = 0"},
		}

		// admin api key
		// org api key (by organization id filter)
		// org api key (by application id filter)
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "ak.organization_id = $2", "$3 = 0"},
			{"ak.id = $1", "a.id = $3", "$2 = 0"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, organizationID, applicationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, organizationID, applicationID)
		default:
			return false, nil
		}
	}
}

// ValidateDeviceProfileAccess validates if the client has access to the
// given device-profile.
func ValidateDeviceProfileAccess(flag Flag, id uuid.UUID) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on o.id = ou.organization_id
		left join application a
			on a.organization_id = o.id
		left join device_profile dp
			on dp.organization_id = o.id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.application_id = a.id
		left join organization o
			on ak.organization_id = o.id
		left join device_profile dp
			on a.organization_id = dp.organization_id or o.id = dp.organization_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// gloabal admin
		// organization users
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "dp.device_profile_id = $2"},
		}

		// admin api key
		// org api key
		// app api key within the same org as the device-profile
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "dp.device_profile_id = $2"}, // dp is joined both on application and organization
		}
	case Update, Delete:
		// global admin
		// organization admin users
		// organization device admin users
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin=true", "dp.device_profile_id = $2"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_device_admin=true", "dp.device_profile_id = $2"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "dp.device_profile_id = $2", "ak.organization_id = dp.organization_id"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, id, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, id)
		default:
			return false, nil
		}
	}
}

// ValidateMulticastGroupsAccess validates if the client has access to the
// multicast-groups.
func ValidateMulticastGroupsAccess(flag Flag, applicationID int64) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join application a
			on a.organization_id = ou.organization_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join application a
			on ak.organization_id = a.organization_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create:
		// global admin
		// organization admin
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "a.id = $2", "ou.is_admin = true"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"},
		}

	case List:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "a.id = $2"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "a.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, applicationID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, applicationID)
		default:
			return false, nil
		}
	}
}

// ValidateMulticastGroupAccess validates if the client has access to the given
// multicast-group.
func ValidateMulticastGroupAccess(flag Flag, multicastGroupID uuid.UUID) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join application a
			on a.organization_id = ou.organization_id
		left join multicast_group mg
			on a.id = mg.application_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join organization o
			on ak.organization_id = o.id
		left join application a
			on o.id = a.organization_id
		left join multicast_group mg
			on a.id = mg.application_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Read:
		// global admin
		// organization users
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "mg.id = $2"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "mg.id = $2"},
		}
	case Update, Delete:
		// global admin
		// organization admin users
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "mg.id = $2"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "mg.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, multicastGroupID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, multicastGroupID)
		default:
			return false, nil
		}
	}
}

// ValidateMulticastGroupQueueAccess validates if the client has access to
// the given multicast-group queue.
func ValidateMulticastGroupQueueAccess(flag Flag, multicastGroupID uuid.UUID) ValidatorFunc {
	userQuery := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join application a
			on a.organization_id = ou.organization_id
		left join multicast_group mg
			on a.id = mg.application_id
	`

	apiKeyQuery := `
		select
			1
		from
			api_key ak
		left join organization o
			on ak.organization_id = o.id
		left join application a
			on o.id = a.organization_id
		left join multicast_group mg
			on a.id = mg.application_id
	`

	var userWhere = [][]string{}
	var apiKeyWhere = [][]string{}

	switch flag {
	case Create, Read, List, Delete:
		// global admin
		// organization user
		userWhere = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "mg.id = $2"},
		}

		// admin api key
		// org api key
		apiKeyWhere = [][]string{
			{"ak.id = $1", "ak.is_admin = true"},
			{"ak.id = $1", "mg.id = $2"},
		}
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		switch claims.Subject {
		case SubjectUser:
			return executeQuery(db, userQuery, userWhere, claims.Username, multicastGroupID, claims.UserID)
		case SubjectAPIKey:
			return executeQuery(db, apiKeyQuery, apiKeyWhere, claims.APIKeyID, multicastGroupID)
		default:
			return false, nil
		}
	}
}

// ValidateAPIKeysAccess validates if the client has access to the global
// API key resource.
func ValidateAPIKeysAccess(flag Flag, organizationID int64, applicationID int64) ValidatorFunc {
	query := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on ou.organization_id = o.id
		left join application a
			on o.id = a.organization_id
	`

	var where [][]string

	switch flag {
	case Create:
		// global admin
		// organization admin of given org id
		// organization admin of given app id
		where = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "ou.is_admin = true", "$2 > 0", "$3 = 0", "o.id = $2"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "ou.is_admin = true", "$3 > 0", "$2 = 0", "a.id = $3"},
		}

	case List:
		// global admin
		// organization admin of given org id (api key filtered by org in api)
		// organization admin of given app id (api key filtered by app in api)
		where = [][]string{
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "ou.is_admin = true", "$2 > 0", "$3 = 0", "o.id = $2"},
			{"(u.email = $1 or u.id = $4)", "u.is_active = true", "ou.is_admin = true", "$3 > 0", "$2 = 0", "a.id = $3"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, query, where, claims.Username, organizationID, applicationID, claims.UserID)
	}
}

// ValidateAPIKeyAccess validates if the client has access to the given API
// key.
func ValidateAPIKeyAccess(flag Flag, id uuid.UUID) ValidatorFunc {
	query := `
		select
			1
		from
			"user" u
		left join organization_user ou
			on u.id = ou.user_id
		left join organization o
			on ou.organization_id = o.id
		left join application a
			on o.id = a.organization_id
		left join api_key ak
			on a.id = ak.application_id or o.id = ak.organization_id or u.is_admin
	`

	var where [][]string
	switch flag {
	case Delete:
		// global admin
		// organization admin
		where = [][]string{
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "u.is_admin = true"},
			{"(u.email = $1 or u.id = $3)", "u.is_active = true", "ou.is_admin = true", "ak.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db sqlx.Queryer, claims *Claims) (bool, error) {
		return executeQuery(db, query, where, claims.Username, id, claims.UserID)
	}
}

func executeQuery(db sqlx.Queryer, query string, where [][]string, args ...interface{}) (bool, error) {
	var ors []string
	for _, ands := range where {
		ors = append(ors, "(("+strings.Join(ands, ") and (")+"))")
	}
	whereStr := strings.Join(ors, " or ")
	query = "select count(*) from (" + query + " where " + whereStr + " limit 1) count_only"

	var count int64
	if err := sqlx.Get(db, &count, query, args...); err != nil {
		return false, errors.Wrap(err, "select error")
	}
	return count > 0, nil
}
