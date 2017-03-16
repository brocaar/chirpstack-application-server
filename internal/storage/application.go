package storage

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var applicationNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Application represents an application.
type Application struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`

	IsABP              bool     `db:"is_abp"`
	IsClassC           bool     `db:"is_class_c"`
	RelaxFCnt          bool     `db:"relax_fcnt"`
	RXWindow           RXWindow `db:"rx_window"`
	RXDelay            uint8    `db:"rx_delay"`
	RX1DROffset        uint8    `db:"rx1_dr_offset"`
	RX2DR              uint8    `db:"rx2_dr"`
	ChannelListID      *int64   `db:"channel_list_id"`
	ADRInterval        uint32   `db:"adr_interval"`
	InstallationMargin float64  `db:"installation_margin"`
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

	if a.RXDelay > 15 {
		return errors.New("max value of RXDelay is 15")
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
			rx_delay,
			rx1_dr_offset,
			channel_list_id,
			rx_window,
			rx2_dr,
			relax_fcnt,
			adr_interval,
			installation_margin,
			is_abp,
			is_class_c
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) returning id`,
		item.Name,
		item.Description,
		item.RXDelay,
		item.RX1DROffset,
		item.ChannelListID,
		item.RXWindow,
		item.RX2DR,
		item.RelaxFCnt,
		item.ADRInterval,
		item.InstallationMargin,
		item.IsABP,
		item.IsClassC,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			default:
				return errors.Wrap(err, "insert error")
			}
		default:
			return errors.Wrap(err, "insert error")
		}
	}
	log.WithFields(log.Fields{
		"id":   item.ID,
		"name": item.Name,
	}).Info("application created")
	return nil
}

// GetApplication returns the Application for the given id.
func GetApplication(db *sqlx.DB, id int64) (Application, error) {
	var app Application
	err := db.Get(&app, "select * from application where id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return app, ErrDoesNotExist
		}
		return app, errors.Wrap(err, "select error")
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
func GetApplicationCountForUser(db *sqlx.DB, username string) (int, error) {
	var count int
	err := db.Get(&count, `
		select
			count(a.*)
		from application a
		inner join application_user au
			on a.id = au.application_id
		inner join "user" u
			on au.user_id = u.id
		where
			u.username = $1
			and u.is_active = true
	`, username)
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
func GetApplicationsForUser(db *sqlx.DB, username string, limit, offset int) ([]Application, error) {
	var apps []Application
	err := db.Select(&apps, `
		select a.*
		from application a
		inner join application_user au
			on a.id = au.application_id
		inner join "user" u
			on au.user_id = u.id
		where
			u.username = $1
			and u.is_active = true
	`, username)
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
			rx_delay = $4,
			rx1_dr_offset = $5,
			channel_list_id = $6,
			rx_window = $7,
			rx2_dr = $8,
			relax_fcnt = $9,
			adr_interval = $10,
			installation_margin = $11,
			is_abp = $12,
			is_class_c = $13
		where id = $1`,
		item.ID,
		item.Name,
		item.Description,
		item.RXDelay,
		item.RX1DROffset,
		item.ChannelListID,
		item.RXWindow,
		item.RX2DR,
		item.RelaxFCnt,
		item.ADRInterval,
		item.InstallationMargin,
		item.IsABP,
		item.IsClassC,
	)
	if err != nil {
		switch err := err.(type) {
		case *pq.Error:
			switch err.Code.Name() {
			case "unique_violation":
				return ErrAlreadyExists
			default:
				return errors.Wrap(err, "update error")
			}
		default:
			return errors.Wrap(err, "update error")
		}
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}

	// update node settings for nodes using the application settings
	_, err = db.Exec(`
		update node
		set
			rx_delay = $2,
			rx1_dr_offset = $3,
			channel_list_id = $4,
			rx_window = $5,
			rx2_dr = $6,
			relax_fcnt = $7,
			adr_interval = $8,
			installation_margin = $9,
			is_abp = $10,
			is_class_c = $11
		where application_id = $1
		and use_application_settings = true`,
		item.ID,
		item.RXDelay,
		item.RX1DROffset,
		item.ChannelListID,
		item.RXWindow,
		item.RX2DR,
		item.RelaxFCnt,
		item.ADRInterval,
		item.InstallationMargin,
		item.IsABP,
		item.IsClassC,
	)
	if err != nil {
		return fmt.Errorf("update nodes error: %s", err)
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

// GetApplicationUsers gets the number of users that have rights to the
// application.
func GetApplicationUsersCount(db *sqlx.DB, applicationID int64) (int32, error) {
	var count int32
	err := db.Get(&count, "select count(*) from application_user where application_id = $1", applicationID)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetUsersForApplication gets the information for the user that has rights to
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
func CreateUserForApplication(db *sqlx.DB, applicationID, userID int64, adminAccess bool) error {
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
