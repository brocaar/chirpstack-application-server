package storage

import (
	"errors"
	"fmt"
	"regexp"

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

// GetApplicationByName returns the Application for the given application name.
func GetApplicationByName(db *sqlx.DB, name string) (Application, error) {
	var app Application
	err := db.Get(&app, "select * from application where name = $1", name)
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

// DeleteApplicationByName deletes the Application matching the given name.
func DeleteApplicationByname(db *sqlx.DB, name string) error {
	res, err := db.Exec("delete from application where name = $1", name)
	if err != nil {
		return fmt.Errorf("delete application error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("application with name %s does not exist", name)
	}
	log.WithFields(log.Fields{
		"name": name,
	}).Info("application deleted")

	return nil
}
