package storage

import (
	"database/sql/driver"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"

	"github.com/brocaar/lorawan"
)

// UsedDevNonceCount is the number of used dev-nonces to track.
const UsedDevNonceCount = 10

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
	Name          string            `db:"name" json:"name"`
	DevEUI        lorawan.EUI64     `db:"dev_eui" json:"devEUI"`
	AppEUI        lorawan.EUI64     `db:"app_eui" json:"appEUI"`
	AppKey        lorawan.AES128Key `db:"app_key" json:"appKey"`
	UsedDevNonces DevNonceList      `db:"used_dev_nonces" json:"usedDevNonces"`

	RXWindow      RXWindow `db:"rx_window" json:"rxWindow"`
	RXDelay       uint8    `db:"rx_delay" json:"rxDelay"`
	RX1DROffset   uint8    `db:"rx1_dr_offset" json:"rx1DROffset"`
	RX2DR         uint8    `db:"rx2_dr" json:"rx2DR"`
	ChannelListID *int64   `db:"channel_list_id" json:"channelListID"`
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
	if len(n.UsedDevNonces) > UsedDevNonceCount {
		n.UsedDevNonces = n.UsedDevNonces[len(n.UsedDevNonces)-UsedDevNonceCount:]
	}

	return true
}

// CreateNode creates the given Node.
func CreateNode(db *sqlx.DB, n Node) error {
	if n.RXDelay > 15 {
		return errors.New("max value of RXDelay is 15")
	}

	_, err := db.Exec(`
		insert into node (
			dev_eui,
			app_eui,
			app_key,
			rx_delay,
			rx1_dr_offset,
			rx_window,
			rx2_dr,
			channel_list_id
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		n.DevEUI[:],
		n.AppEUI[:],
		n.AppKey[:],
		n.RXDelay,
		n.RX1DROffset,
		n.RXWindow,
		n.RX2DR,
		n.ChannelListID,
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
			app_eui = $2,
			app_key = $3,
			used_dev_nonces = $4,
			rx_delay = $5,
			rx1_dr_offset = $6,
			rx_window = $7,
			rx2_dr = $8,
			channel_list_id = $9
		where dev_eui = $1`,
		n.DevEUI[:],
		n.AppEUI[:],
		n.AppKey[:],
		n.UsedDevNonces,
		n.RXDelay,
		n.RX1DROffset,
		n.RXWindow,
		n.RX2DR,
		n.ChannelListID,
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
// TODO: remove node-session
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

	channels, err := GetChannelsForChannelList(db, *node.ChannelListID)
	if err != nil {
		return nil, err
	}

	var cFList lorawan.CFList
	for _, channel := range channels {
		if len(cFList) <= channel.Channel-3 {
			return nil, fmt.Errorf("invalid channel index for CFList: %d", channel.Channel)
		}
		cFList[channel.Channel-3] = uint32(channel.Frequency)
	}
	return &cFList, nil
}
