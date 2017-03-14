package storage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/brocaar/lorawan"
)

var nodeNameRegexp = regexp.MustCompile(`^[\w-]+$`)

// DevNonceList represents a list of dev nonces
type DevNonceList [][2]byte

// Scan implements the sql.Scanner interface.
func (l *DevNonceList) Scan(src interface{}) error {
	if src == nil {
		*l = make([][2]byte, 0)
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("src must be of type []byte, got: %T", src)
	}
	if len(b)%2 != 0 {
		return errors.New("the length of src must be a multiple of 2")
	}
	for i := 0; i < len(b); i += 2 {
		*l = append(*l, [2]byte{b[i], b[i+1]})
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (l DevNonceList) Value() (driver.Value, error) {
	b := make([]byte, 0, len(l)/2)
	for _, n := range l {
		b = append(b, n[:]...)
	}
	return b, nil
}

// RXWindow defines the RX window option.
type RXWindow int8

// Available RX window options.
const (
	RX1 = iota
	RX2
)

// Scan implements the sql.Scanner interface.
func (r *RXWindow) Scan(src interface{}) error {
	i, ok := src.(int64)
	if !ok {
		return fmt.Errorf("expected int64, got: %T", src)
	}
	*r = RXWindow(i)
	return nil
}

// Value implements the driver.Valuer interface.
func (r RXWindow) Value() (driver.Value, error) {
	return int64(r), nil
}

// Node contains the information of a node.
type Node struct {
	ApplicationID          int64             `db:"application_id"`
	UseApplicationSettings bool              `db:"use_application_settings"`
	Name                   string            `db:"name"`
	Description            string            `db:"description"`
	DevEUI                 lorawan.EUI64     `db:"dev_eui"`
	AppEUI                 lorawan.EUI64     `db:"app_eui"`
	AppKey                 lorawan.AES128Key `db:"app_key"`
	IsABP                  bool              `db:"is_abp"`
	IsClassC               bool              `db:"is_class_c"`
	DevAddr                lorawan.DevAddr   `db:"dev_addr"`
	NwkSKey                lorawan.AES128Key `db:"nwk_s_key"`
	AppSKey                lorawan.AES128Key `db:"app_s_key"`
	UsedDevNonces          DevNonceList      `db:"used_dev_nonces"`
	RelaxFCnt              bool              `db:"relax_fcnt"`

	RXWindow      RXWindow `db:"rx_window"`
	RXDelay       uint8    `db:"rx_delay"`
	RX1DROffset   uint8    `db:"rx1_dr_offset"`
	RX2DR         uint8    `db:"rx2_dr"`
	ChannelListID *int64   `db:"channel_list_id"`

	ADRInterval        uint32  `db:"adr_interval"`
	InstallationMargin float64 `db:"installation_margin"`
}

// Validate validates the data of the Node.
func (n Node) Validate() error {
	if !nodeNameRegexp.MatchString(n.Name) {
		return ErrNodeInvalidName
	}
	if n.RXDelay > 15 {
		return ErrNodeMaxRXDelay
	}

	return nil
}

// ValidateDevNonce returns if the given dev-nonce is valid.
// When valid, it will be added to UsedDevNonces. This does
// not update the Node in the database!
func (n *Node) ValidateDevNonce(nonce [2]byte) bool {
	for _, used := range n.UsedDevNonces {
		if nonce == used {
			return false
		}
	}
	n.UsedDevNonces = append(n.UsedDevNonces, nonce)
	return true
}

func updateNodeSettingsFromApplication(db *sqlx.DB, n *Node) error {
	app, err := GetApplication(db, n.ApplicationID)
	if err != nil {
		return fmt.Errorf("get application error: %s", err)
	}

	n.RXDelay = app.RXDelay
	n.RX1DROffset = app.RX1DROffset
	n.ChannelListID = app.ChannelListID
	n.RXWindow = app.RXWindow
	n.RX2DR = app.RX2DR
	n.RelaxFCnt = app.RelaxFCnt
	n.ADRInterval = app.ADRInterval
	n.InstallationMargin = app.InstallationMargin
	n.IsABP = app.IsABP
	n.IsClassC = app.IsClassC

	return nil
}

// CreateNode creates the given Node.
func CreateNode(db *sqlx.DB, n Node) error {
	if err := n.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	if n.UseApplicationSettings {
		if err := updateNodeSettingsFromApplication(db, &n); err != nil {
			return err
		}
	}

	_, err := db.Exec(`
		insert into node (
			application_id,
			name,
			description,
			dev_eui,
			app_eui,
			app_key,
			dev_addr,
			app_s_key,
			nwk_s_key,
			rx_delay,
			rx1_dr_offset,
			rx_window,
			rx2_dr,
			channel_list_id,
			relax_fcnt,
			adr_interval,
			installation_margin,
			is_abp,
			is_class_c,
			use_application_settings
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`,
		n.ApplicationID,
		n.Name,
		n.Description,
		n.DevEUI[:],
		n.AppEUI[:],
		n.AppKey[:],
		n.DevAddr[:],
		n.AppSKey[:],
		n.NwkSKey[:],
		n.RXDelay,
		n.RX1DROffset,
		n.RXWindow,
		n.RX2DR,
		n.ChannelListID,
		n.RelaxFCnt,
		n.ADRInterval,
		n.InstallationMargin,
		n.IsABP,
		n.IsClassC,
		n.UseApplicationSettings,
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
	log.WithField("dev_eui", n.DevEUI).Info("node created")
	return nil
}

// UpdateNode updates the given Node.
// TODO: change node into pointer
func UpdateNode(db *sqlx.DB, n Node) error {
	if err := n.Validate(); err != nil {
		return errors.Wrap(err, "validate error")
	}

	if n.UseApplicationSettings {
		if err := updateNodeSettingsFromApplication(db, &n); err != nil {
			return err
		}
	}

	res, err := db.Exec(`
		update node set
			application_id = $2,
			name = $3,
			description = $4,
			app_eui = $5,
			app_key = $6,
			dev_addr = $7,
			app_s_key = $8,
			nwk_s_key = $9,
			used_dev_nonces = $10,
			rx_delay = $11,
			rx1_dr_offset = $12,
			rx_window = $13,
			rx2_dr = $14,
			channel_list_id = $15,
			relax_fcnt = $16,
			adr_interval = $17,
			installation_margin = $18,
			is_abp = $19,
			is_class_c = $20,
			use_application_settings = $21
		where dev_eui = $1`,
		n.DevEUI[:],
		n.ApplicationID,
		n.Name,
		n.Description,
		n.AppEUI[:],
		n.AppKey[:],
		n.DevAddr[:],
		n.AppSKey[:],
		n.NwkSKey[:],
		n.UsedDevNonces,
		n.RXDelay,
		n.RX1DROffset,
		n.RXWindow,
		n.RX2DR,
		n.ChannelListID,
		n.RelaxFCnt,
		n.ADRInterval,
		n.InstallationMargin,
		n.IsABP,
		n.IsClassC,
		n.UseApplicationSettings,
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
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get rows affected error")
	}
	if ra == 0 {
		return ErrDoesNotExist
	}
	log.WithField("dev_eui", n.DevEUI).Info("node updated")
	return nil
}

// DeleteNode deletes the Node matching the given DevEUI.
func DeleteNode(db *sqlx.DB, devEUI lorawan.EUI64) error {
	res, err := db.Exec("delete from node where dev_eui = $1",
		devEUI[:],
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
	log.WithField("dev_eui", devEUI).Info("node deleted")
	return nil
}

// GetNode returns the Node for the given DevEUI.
func GetNode(db *sqlx.DB, devEUI lorawan.EUI64) (Node, error) {
	var node Node
	err := db.Get(&node, "select * from node where dev_eui = $1", devEUI[:])
	if err != nil {
		if err == sql.ErrNoRows {
			return node, ErrDoesNotExist
		}
		return node, errors.Wrap(err, "select error")
	}
	return node, nil
}

// GetNodesForApplicationID returns a slice of nodes for the given application
// id, sorted by name.
func GetNodesForApplicationID(db *sqlx.DB, applicationID int64, limit, offset int) ([]Node, error) {
	var nodes []Node
	err := db.Select(&nodes, `
		select *
		from node
		where
			application_id = $1
		order by name
		limit $2 offset $3`,
		applicationID,
		limit,
		offset,
	)
	if err != nil {
		return nodes, errors.Wrap(err, "select error")
	}
	return nodes, nil
}

// GetNodesCountForApplicationID returns the total number of nodes for the
// given applicaiton id.
func GetNodesCountForApplicationID(db *sqlx.DB, applicationID int64) (int, error) {
	var count int
	err := db.Get(&count, `
		select count(*)
		from node
		where
			application_id = $1`,
		applicationID)
	if err != nil {
		return 0, errors.Wrap(err, "select error")
	}
	return count, nil
}

// GetCFListForNode returns the CFList for the given node if the
// used ISM band allows using a CFList.
func GetCFListForNode(db *sqlx.DB, node Node) (*lorawan.CFList, error) {
	if node.ChannelListID == nil {
		return nil, nil
	}

	var cFList lorawan.CFList
	cl, err := GetChannelList(db, *node.ChannelListID)
	if err != nil {
		return nil, errors.Wrap(err, "get channel list error")
	}

	if len(cl.Channels) > len(cFList) {
		return nil, ErrCFListTooManyChannels
	}

	for i, v := range cl.Channels {
		cFList[i] = uint32(v)
	}
	return &cFList, nil
}
