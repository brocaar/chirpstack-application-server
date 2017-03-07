package storage

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
)

var applicationNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// Application represents an application.
type Application struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

// UserAccess represents the users that have access to an application
type UserAccess struct {
	UserId    int64     `db:"user_id"`
	Username  string    `db:"username"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate validates the data of the Application.
func (a Application) Validate() error {
	if !applicationNameRegexp.MatchString(a.Name) {
		return errors.New("application name may only contain words, numbers and dashes")
	}
	return nil
}

// CreateApplication creates the given Application.
func CreateApplication(db *sqlx.DB, item *Application) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("validate application error: %s", err)
	}

	err := db.Get(&item.ID, `
		insert into application (
			name,
			description
		) values ($1, $2) returning id`,
		item.Name,
		item.Description,
	)
	if err != nil {
		return fmt.Errorf("create application error: %s", err)
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
		return app, fmt.Errorf("get application error: %s", err)
	}
	return app, nil
}

// GetApplicationCount returns the total number of applications.
func GetApplicationCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, "select count(*) from application")
	if err != nil {
		return 0, fmt.Errorf("get applications count error: %s", err)
	}
	return count, nil
}

// GetApplications returns a slice of applications, sorted by name and
// respecting the given limit and offset.
func GetApplications(db *sqlx.DB, limit, offset int) ([]Application, error) {
	var apps []Application
	err := db.Select(&apps, "select * from application order by name limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get applications error: %s", err)
	}
	return apps, nil
}

// UpdateApplication updates the given Application.
func UpdateApplication(db *sqlx.DB, item Application) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf("validate application error: %s", err)
	}

	res, err := db.Exec(`
		update application
		set
			name = $2,
			description = $3
		where id = $1`,
		item.ID,
		item.Name,
		item.Description,
	)
	if err != nil {
		return fmt.Errorf("update application error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("application %d does not exist", item.ID)
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
		return fmt.Errorf("delete application error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("application with id %d does not exist", id)
	}
	log.WithFields(log.Fields{
		"id": id,
	}).Info("application deleted")

	return nil
}


// GetApplicationUsers lists the users that have rights to the application, 
// given the offset into the list and the number of users to return.
func GetApplicationUsers(db *sqlx.DB, applicationId int64, limit, offset int) ([]UserAccess, error) {
	var users []UserAccess
	err := db.Select(&users, `select au.user_id as user_id, 
	                                 au.is_admin as is_admin, 
	                                 au.created_at as created_at, 
	                                 au.updated_at as updated_at,
	                                 u.username as username 
	                          from application_user au, "user" as u 
	                          where au.application_id = $1 and au.user_id = u.id 
	                          order by user_id limit $2 offset $3`,
	                          applicationId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get users for application error: %s", err)
	}
	return users, nil
}

// GetApplicationUsers gets the number of users that have rights to the 
// application.
func GetApplicationUsersCount(db *sqlx.DB, applicationId int64) (int32, error) {
	var count int32
	err := db.Get(&count, "select count(*) from application_user where application_id = $1", applicationId)
	if err != nil {
		return 0, fmt.Errorf("get user count for application error: %s", err)
	}
	return count, nil
}

// GetUsersForApplication gets the information for the user that has rights to 
// the application.
func GetUserForApplication(db *sqlx.DB, applicationId, userId int64) (*UserAccess, error) {
	var user UserAccess
	err := db.Get(&user, `select au.user_id as user_id, 
	                             au.is_admin as is_admin, 
	                             au.created_at as created_at, 
	                             au.updated_at as updated_at,
	                             u.username as username 
	                          from application_user au, "user" as u 
	                          where au.application_id = $1 and au.user_id = $2 and au.user_id = u.id and user_id = $2`, 
	                          applicationId, userId)
	if err != nil {
		return nil, fmt.Errorf("get user for application error: %s", err)
	}
	return &user, nil
}

// CreateUserForApplication adds the user to the application with the given
// access.
func CreateUserForApplication(db *sqlx.DB, applicationId, userId int64, adminAccess bool ) error {
	// Add the new user.
	rows, err := db.Queryx( `
		insert into application_user (
			application_id,
			user_id,
			is_admin,
			created_at,
			updated_at
		) values ($1, $2, $3, now(), now())`,
		applicationId,
		userId,
		adminAccess,
	)
    if err != nil {
    	// Unexpected error
    	return err
    }
    rows.Close()
	log.WithFields(log.Fields{
		"userId": userId,
		"applicationId":   applicationId,
		"admin": adminAccess,
	}).Info("user for application created")
	return nil
}

// UpdateUserForApplication lets the caller update the admin setting for the 
// user for the application.
func UpdateUserForApplication(db *sqlx.DB, applicationId, userId int64, adminAccess bool ) error {
	rows, err := db.Queryx( "update application_user set is_admin = $1, updated_at = now() where application_id = $2 and user_id = $3",
		adminAccess,
		applicationId,
		userId,
	)
    rows.Close()
    if err != nil {
    	// Unexpected error
    	return err
    }
	log.WithFields(log.Fields{
		"userId": userId,
		"applicationId":   applicationId,
		"admin": adminAccess,
	}).Info("user for application updated")
	return nil
}

// DeleteUserForApplication lets the caller remove the user from the application.
func DeleteUserForApplication(db *sqlx.DB, applicationId, userId int64) error {
	rows, err := db.Queryx( "delete from application_user where application_id = $1 and user_id = $2",
		applicationId,
		userId,
	)
    rows.Close()
    if err != nil {
    	// Unexpected error
    	return err
    }
	log.WithFields(log.Fields{
		"userId": userId,
		"applicationId":   applicationId,
	}).Info("user for application deleted")
	return nil
}