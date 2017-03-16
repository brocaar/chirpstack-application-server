package auth

import (
	"strings"

	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
)

const userQuery = `select count(*) from "user" u left join application_user au on u.id = au.user_id left join application a on au.application_id = a.id left join node n on a.id = n.application_id`

// ValidateActiveUser validates if the user in the JWT claim is active.
func ValidateActiveUser() ValidatorFunc {
	where := [][]string{
		{"u.username = $1", "u.is_active = true"},
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateUsersAccess validates if the client has access to the global users
// resource.
func ValidateUsersAccess(flag Flag) ValidatorFunc {
	var where [][]string

	switch flag {
	case Create, List:
		// global admin or admin within an application
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateUserAccess validates if the client has access to the given user
// resource.
func ValidateUserAccess(userID int64, flag Flag) ValidatorFunc {
	var where [][]string

	switch flag {
	case Read:
		// global admin or user itself
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
		// global admin and user itself
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "u.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, userID)
	}
}

// ValidateApplicationsAccess validates if the client has access to the
// global applications resource.
func ValidateApplicationsAccess(flag Flag) ValidatorFunc {
	var where [][]string

	switch flag {
	case Create:
		// global admin user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
		}
	case List:
		// all users (output will be based on authenticated user)
		where = [][]string{
			{"u.username = $1", "u.is_active = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

// ValidateApplicationAccess validates if the client has access to the given
// application.
func ValidateApplicationAccess(applicationID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin users or users assigned to application
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "a.id = $2"},
		}
	case Update:
		// global admin or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "$2 = $2"},
		}
	case Delete:
		// global admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$2 = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateApplicationMembersAccess validates if the client has access to the
// given application members.
func ValidateApplicationMembersAccess(applicationID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "a.id = $2"},
		}
	case List:
		// global admin or application user
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateApplicationMemberAccess validates if the client has access to the
// given application member.
func ValidateApplicationMemberAccess(applicationID, userID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin, application admin or user itself.
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "a.id = $2"},
			{"u.username = $1", "u.is_active = true", "a.id = $2", "au.user_id = $3", "au.user_id = u.id"},
		}
	case Update:
		// global admin or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "a.id = $2"},
		}
	case Delete:
		// global admin or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true", "$3 = $3"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID, userID)
	}
}

// ValidateNodesAccess validates if the client has access to the global nodes
// resource.
func ValidateNodesAccess(applicationID int64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create:
		// global admin user or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "a.id = $2"},
		}
	case List:
		// global amdin user or users assigned to application
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "a.id = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, applicationID)
	}
}

// ValidateNodeAccess validates if the client has access to the given node.
func ValidateNodeAccess(devEUI lorawan.EUI64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Read:
		// global admin user or users assigned to application
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "n.dev_eui = $2"},
		}
	case Update:
		// global admin user or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "n.dev_eui = $2"},
		}
	case Delete:
		// global admin user or application admin
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "au.is_admin = true", "n.dev_eui = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, devEUI[:])
	}
}

// ValidateNodeQueueAccess validates if the client has access to the queue
// of the given node.
func ValidateNodeQueueAccess(devEUI lorawan.EUI64, flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create, Read, List, Update, Delete:
		// global admin users or application users
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
			{"u.username = $1", "u.is_active = true", "n.dev_eui = $2"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username, devEUI[:])
	}
}

// ValidateChannelListAccess validates if the client has access to the channel-lists.
func ValidateChannelListAccess(flag Flag) ValidatorFunc {
	var where = [][]string{}

	switch flag {
	case Create, Update, Delete:
		// global admin users
		where = [][]string{
			{"u.username = $1", "u.is_active = true", "u.is_admin = true"},
		}
	case Read, List:
		// any user
		where = [][]string{
			{"u.username = $1", "u.is_active = true"},
		}
	default:
		panic("unsupported flag")
	}

	return func(db *sqlx.DB, claims *Claims) (bool, error) {
		return executeQuery(db, userQuery, where, claims.Username)
	}
}

func executeQuery(db *sqlx.DB, query string, where [][]string, args ...interface{}) (bool, error) {
	var ors []string
	for _, ands := range where {
		ors = append(ors, "("+strings.Join(ands, " and ")+")")
	}
	whereStr := strings.Join(ors, " or ")
	query = query + " where " + whereStr

	var count int64
	if err := db.Get(&count, query, args...); err != nil {
		return false, errors.Wrap(err, "select error")
	}
	return count > 0, nil
}
