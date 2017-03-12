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

// Validate validates the data of the Application.
func (a Application) Validate() error {
	if !applicationNameRegexp.MatchString(a.Name) {
		return errors.New("application name may only contain words, numbers and dashes")
	}

	if a.RXDelay > 15 {
		return errors.New("max value of RXDelay is 15")
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
		return fmt.Errorf("update application error: %s", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("application %d does not exist", item.ID)
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
