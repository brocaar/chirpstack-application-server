package code

import (
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type multicastGroupCount struct {
	ID    uuid.UUID `db:"id"`
	Name  string    `db:"name"`
	Count int64     `db:"count"`
}

// Validate that multicast-group devices are under a single application.
func ValidateMulticastGroupDevices(db sqlx.Ext) error {
	var items []multicastGroupCount

	err := sqlx.Select(db, &items, `
		select
			mg.id,
			mg.name,
			count(distinct d.application_id) as count
		from
			multicast_group mg
		inner join device_multicast_group dmg
			on dmg.multicast_group_id = mg.id
		inner join device d
			on d.dev_eui = dmg.dev_eui
		group by
			mg.id
	`)
	if err != nil {
		return errors.Wrap(err, "select error")
	}

	for _, item := range items {
		if item.Count != 1 {
			log.WithFields(log.Fields{
				"multicast_group_id":   item.ID,
				"multicast_group_name": item.Name,
			}).Fatal("Multicast-group contains devices from multiple applications. Please read the changelog why you are seeing this error.")
		}
	}

	err = sqlx.Select(db, &items, `
		select
			mg.id,
			mg.name,
			count(distinct dmg.dev_eui) as count
		from
			multicast_group mg
		left join device_multicast_group dmg
			on mg.id = dmg.multicast_group_id
		group by
			mg.id
	`)
	if err != nil {
		return errors.Wrap(err, "select error")
	}

	for _, item := range items {
		if item.Count == 0 {
			log.WithFields(log.Fields{
				"multicast_group_id":   item.ID,
				"multicast_group_name": item.Name,
			}).Fatal("Multicast-group does not contain any devices. Please read the changelog why you are seeing this error.")
		}
	}

	return nil
}
