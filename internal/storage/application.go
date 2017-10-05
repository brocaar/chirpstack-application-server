package storage

import (
	"fmt"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var applicationNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Application represents an application.
type Application struct {
	ID               int64  `db:"id"`
	Name             string `db:"name"`
	Description      string `db:"description"`
	OrganizationID   int64  `db:"organization_id"`
	ServiceProfileID string `db:"service_profile_id"`
}

// UserAccess represents the users that have access to an application
type UserAccess struct {
	UserID    int64     `db:"user_id"`
	Username  string    `db:"username"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate validates the data of the Application.
func (a Application) Validate() error {
	if !applicationNameRegexp.MatchString(a.Name) {
		return ErrApplicationInvalidName
	}

	return nil
}

// CreateApplication creates the given Application.
func CreateApplication(db *sqlx.DB, item *Application) error {
	if err := item.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	err := db.Get(&item.ID, `
		insert into application (
			name,
			description,
			organization_id,
			service_profile_id
		) values ($1, $2, $3, $4) returning id`,
		item.Name,
		item.Description,
		item.OrganizationID,
		item.ServiceProfileID,
	)
	if err != nil {
		return handlePSQLError(err, "insert error")
	}

	log.WithFields(log.Fields{
		"id":   item.ID,
		"name": item.Name,
	}).Info("application created")

	return nil
}

// GetApplication returns the Application for the given id.
func GetApplication(db sqlx.Queryer, id int64) (Application, error) {
	var app Application
	err := sqlx.Get(db, &app, "select * from application where id = $1", id)
	if err != nil {
		return app, handlePSQLError(err, "select error")
	}

	return app, nil
}

// GetApplicationCount returns the total number of applications.
func GetApplicationCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, "select count(*) from application")
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetApplicationCountForUser returns the total number of applications
// available for the given user.
// When an organizationID is given, the results will be filtered by this
//
func GetApplicationCountForUser(db *sqlx.DB, username string, organizationID int64) (int, error) {
	var count int
	err := db.Get(&count, `
		select
			count(a.*)
		from application a
		left join application_user au
			on a.id = au.application_id
		left join organization_user ou
			on a.organization_id = ou.organization_id
		inner join "user" u
			on au.user_id = u.id or ou.user_id = u.id
		where
			u.username = $1
			and u.is_active = true
			and (au.user_id is null or au.user_id = u.id)
			and (ou.user_id is null or ou.user_id = u.id)
			and (
				$2 = 0
				or a.organization_id = $2
			)
	`, username, organizationID)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetApplicationCountForOrganizationID returns the total number of
// applications for the given organization.
func GetApplicationCountForOrganizationID(db *sqlx.DB, organizationID int64) (int, error) {
	var count int
	err := db.Get(&count, `
		select count(*)
		from application
		where
			organization_id = $1`,
		organizationID,
	)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetApplications returns a slice of applications, sorted by name and
// respecting the given limit and offset.
func GetApplications(db *sqlx.DB, limit, offset int) ([]Application, error) {
	var apps []Application
	err := db.Select(&apps, "select * from application order by name limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return apps, nil
}

// GetApplicationsForUser returns a slice of application of which the given
// user is a member of.
func GetApplicationsForUser(db *sqlx.DB, username string, organizationID int64, limit, offset int) ([]Application, error) {
	var apps []Application
	err := db.Select(&apps, `
		select a.*
		from application a
		left join application_user au
			on a.id = au.application_id
		left join organization_user ou
			on a.organization_id = ou.organization_id
		inner join "user" u
			on au.user_id = u.id or ou.user_id = u.id
		where
			u.username = $1
			and u.is_active = true
			and (au.user_id is null or au.user_id = u.id)
			and (ou.user_id is null or ou.user_id = u.id)
			and (
				$2 = 0
				or a.organization_id = $2
			)
		order by a.name
		limit $3 offset $4
	`, username, organizationID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}

	return apps, nil
}

// GetApplicationsForOrganizationID returns a slice of applications for the given
// organization.
func GetApplicationsForOrganizationID(db *sqlx.DB, organizationID int64, limit, offset int) ([]Application, error) {
	var apps []Application
	err := db.Select(&apps, `
		select *
		from application
		where
			organization_id = $1
		order by name
		limit $2 offset $3`,
		organizationID,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}

	return apps, nil
}

// UpdateApplication updates the given Application.
// When the application contains nodes with UseApplicationSettings=true, this
// will also update these nodes.
func UpdateApplication(db *sqlx.DB, item Application) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("validate application error: %s", err)
	}

	res, err := db.Exec(`
		update application
		set
			name = $2,
			description = $3,
			organization_id = $4,
			service_profile_id = $5
		where id = $1`,
		item.ID,
		item.Name,
		item.Description,
		item.OrganizationID,
		item.ServiceProfileID,
	)
	if err != nil {
		return handlePSQLError(err, "update error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"id":   item.ID,
		"name": item.Name,
	}).Info("application updated")

	return nil
}

// DeleteApplication deletes the Application matching the given ID.
func DeleteApplication(db *sqlx.DB, id int64) error {
	res, err := db.Exec("delete from application where id = $1", id)
	if err != nil {
		return handlePSQLError(err, "delete error")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Info("application deleted")

	return nil
}

// GetApplicationUsers lists the users that have rights to the application,
// given the offset into the list and the number of users to return.
func GetApplicationUsers(db *sqlx.DB, applicationID int64, limit, offset int) ([]UserAccess, error) {
	var users []UserAccess
	err := db.Select(&users, `select au.user_id as user_id,
	                                 au.is_admin as is_admin,
	                                 au.created_at as created_at,
	                                 au.updated_at as updated_at,
	                                 u.username as username
	                          from application_user au, "user" as u
	                          where au.application_id = $1 and au.user_id = u.id
	                          order by user_id limit $2 offset $3`,
		applicationID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return users, nil
}

// GetApplicationUsersCount gets the number of users that have rights to the
// application.
func GetApplicationUsersCount(db *sqlx.DB, applicationID int64) (int32, error) {
	var count int32
	err := db.Get(&count, "select count(*) from application_user where application_id = $1", applicationID)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetUserForApplication gets the information for the user that has rights to
// the application.
func GetUserForApplication(db *sqlx.DB, applicationID, userID int64) (*UserAccess, error) {
	var user UserAccess
	err := db.Get(&user, `select au.user_id as user_id,
	                             au.is_admin as is_admin,
	                             au.created_at as created_at,
	                             au.updated_at as updated_at,
	                             u.username as username
	                          from application_user au, "user" as u
	                          where au.application_id = $1 and au.user_id = $2 and au.user_id = u.id and user_id = $2`,
		applicationID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "select error")
	}
	return &user, nil
}

// CreateUserForApplication adds the user to the application with the given
// access.
func CreateUserForApplication(db sqlx.Execer, applicationID, userID int64, adminAccess bool) error {
	_, err := db.Exec(`
		insert into application_user (
			application_id,
			user_id,
			is_admin,
			created_at,
			updated_at
		) values ($1, $2, $3, now(), now())`,
		applicationID,
		userID,
		adminAccess,
	)

	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			case "foreign_key_violation":
				return ErrDoesNotExist
			default:
				return errors.Wrap(err, "insert error")
			}
		default:
			return errors.Wrap(err, "insert error")
		}
	}

	log.WithFields(log.Fields{
		"user_id":        userID,
		"application_id": applicationID,
		"admin":          adminAccess,
	}).Info("user for application created")
	return nil
}

// UpdateUserForApplication lets the caller update the admin setting for the
// user for the application.
func UpdateUserForApplication(db *sqlx.DB, applicationID, userID int64, adminAccess bool) error {
	_, err := db.Exec("update application_user set is_admin = $1, updated_at = now() where application_id = $2 and user_id = $3",
		adminAccess,
		applicationID,
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "update error")
	}

	log.WithFields(log.Fields{
		"user_id":        userID,
		"application_id": applicationID,
		"admin":          adminAccess,
	}).Info("user for application updated")
	return nil
}

// DeleteUserForApplication lets the caller remove the user from the application.
func DeleteUserForApplication(db *sqlx.DB, applicationID, userID int64) error {
	res, err := db.Exec("delete from application_user where application_id = $1 and user_id = $2",
		applicationID,
		userID,
	)
	if err != nil {
		return errors.Wrap(err, "delete error")
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	log.WithFields(log.Fields{
		"user_id":        userID,
		"application_id": applicationID,
	}).Info("user for application deleted")
	return nil
}
