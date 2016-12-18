package storage

import (
	"database/sql/driver"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"

	"github.com/brocaar/lorawan"
)

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
	Name          string            `db:"name"`
	DevEUI        lorawan.EUI64     `db:"dev_eui"`
	AppEUI        lorawan.EUI64     `db:"app_eui"`
	AppKey        lorawan.AES128Key `db:"app_key"`
	DevAddr       lorawan.DevAddr   `db:"dev_addr"`
	NwkSKey       lorawan.AES128Key `db:"nwk_s_key"`
	AppSKey       lorawan.AES128Key `db:"app_s_key"`
	UsedDevNonces DevNonceList      `db:"used_dev_nonces"`
	RelaxFCnt     bool              `db:"relax_fcnt"`

	RXWindow      RXWindow `db:"rx_window"`
	RXDelay       uint8    `db:"rx_delay"`
	RX1DROffset   uint8    `db:"rx1_dr_offset"`
	RX2DR         uint8    `db:"rx2_dr"`
	ChannelListID *int64   `db:"channel_list_id"`

	ADRInterval        uint32  `db:"adr_interval"`
	InstallationMargin float64 `db:"installation_margin"`
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

// CreateNode creates the given Node.
func CreateNode(db *sqlx.DB, n Node) error {
	if n.RXDelay > 15 {
		return errors.New("max value of RXDelay is 15")
	}

	_, err := db.Exec(`
		insert into node (
			name,
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
			installation_margin
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		n.Name,
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
	)
	if err != nil {
		return fmt.Errorf("create node %s error: %s", n.DevEUI, err)
	}
	log.WithField("dev_eui", n.DevEUI).Info("node created")
	return nil
}

// UpdateNode updates the given Node.
func UpdateNode(db *sqlx.DB, n Node) error {
	if n.RXDelay > 15 {
		return errors.New("max value of RXDelay is 15")
	}

	res, err := db.Exec(`
		update node set
			name = $1,
			app_eui = $2,
			app_key = $3,
			dev_addr = $4,
			app_s_key = $5,
			nwk_s_key = $6,
			used_dev_nonces = $7,
			rx_delay = $8,
			rx1_dr_offset = $9,
			rx_window = $10,
			rx2_dr = $11,
			channel_list_id = $12,
			relax_fcnt = $13,
			adr_interval = $14,
			installation_margin = $15
		where dev_eui = $16`,
		n.Name,
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
		n.DevEUI[:],
	)
	if err != nil {
		return fmt.Errorf("update node %s error: %s", n.DevEUI, err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("node %s does not exist", n.DevEUI)
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
		return fmt.Errorf("delete node %s error: %s", devEUI, err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return fmt.Errorf("node %s does not exist", devEUI)
	}
	log.WithField("dev_eui", devEUI).Info("node deleted")
	return nil
}

// GetNode returns the Node for the given DevEUI.
func GetNode(db *sqlx.DB, devEUI lorawan.EUI64) (Node, error) {
	var node Node
	err := db.Get(&node, "select * from node where dev_eui = $1", devEUI[:])
	if err != nil {
		return node, fmt.Errorf("get node %s error: %s", devEUI, err)
	}
	return node, nil
}

// GetNodesCount returns the total number of nodes.
func GetNodesCount(db *sqlx.DB) (int, error) {
	var count struct {
		Count int
	}
	err := db.Get(&count, "select count(*) as count from node")
	if err != nil {
		return 0, fmt.Errorf("get nodes count error: %s", err)
	}
	return count.Count, nil
}

// GetNodes returns a slice of nodes, sorted by DevEUI.
func GetNodes(db *sqlx.DB, limit, offset int) ([]Node, error) {
	var nodes []Node
	err := db.Select(&nodes, "select * from node order by dev_eui limit $1 offset $2", limit, offset)
	if err != nil {
		return nodes, fmt.Errorf("get nodes error: %s", err)
	}
	return nodes, nil
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
		return nil, err
	}

	if len(cl.Channels) > len(cFList) {
		return nil, errors.New("too many channels in channel-list")
	}

	for i, v := range cl.Channels {
		cFList[i] = uint32(v)
	}
	return &cFList, nil
}
