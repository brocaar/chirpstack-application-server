// Package nsmigrate migrates the old node-session data into PostgreSQL.
package nsmigrate

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lorawan"
	"github.com/garyburd/redigo/redis"
)

const (
	nodeSessionKeyTempl        = "node_session_%s"
	nodeSessionMACTXQueueTempl = "node_session_mac_tx_queue_%s"
)

// NodeSession contains the informatio of a node-session (an activated node).
// This is the old node session.
type NodeSession struct {
	DevAddr  lorawan.DevAddr   `json:"devAddr"`
	AppEUI   lorawan.EUI64     `json:"appEUI"`
	DevEUI   lorawan.EUI64     `json:"devEUI"`
	AppSKey  lorawan.AES128Key `json:"appSKey"`
	NwkSKey  lorawan.AES128Key `json:"nwkSKey"`
	FCntUp   uint32            `json:"fCntUp"`
	FCntDown uint32            `json:"fCntDown"`

	RXWindow    RXWindow `json:"rxWindow"`
	RXDelay     uint8    `json:"rxDelay"`
	RX1DROffset uint8    `json:"rx1DROffset"`
	RX2DR       uint8    `json:"rx2DR"`

	CFList *lorawan.CFList `json:"cFlist"`
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

func getNodeSessionByDevEUI(p *redis.Pool, devEUI lorawan.EUI64) (NodeSession, error) {
	var ns NodeSession

	c := p.Get()
	defer c.Close()

	devAddr, err := redis.String(c.Do("GET", fmt.Sprintf(nodeSessionKeyTempl, devEUI)))
	if err != nil {
		return ns, fmt.Errorf("get node-session pointer for node %s error: %s", devEUI, err)
	}

	val, err := redis.Bytes(c.Do("GET", fmt.Sprintf(nodeSessionKeyTempl, devAddr)))
	if err != nil {
		return ns, fmt.Errorf("get node-session for DevAddr %s error: %s", devAddr, err)
	}

	err = gob.NewDecoder(bytes.NewReader(val)).Decode(&ns)
	if err != nil {
		return ns, fmt.Errorf("decode node-session %s error: %s", devAddr, err)
	}

	return ns, nil
}

// Migrate migrates some of the data from the Redis storage to the node table
// in the PostgreSQL database. This only happens for nodes with a
// blank DevAddr (00000000).
func Migrate(ctx common.Context) {
	var nodes []storage.Node
	if err := ctx.DB.Select(&nodes, "select * from node where dev_addr = $1", []byte{0, 0, 0, 0}); err != nil {
		log.Warningf("storage/nsmigrate: node-session migration failed: %s", err)
	}

	for _, n := range nodes {
		ns, err := getNodeSessionByDevEUI(ctx.RedisPool, n.DevEUI)
		if err != nil {
			log.WithField("dev_eui", n.DevEUI).Warningf("storage/nsmigrate: could not get node-session for DevEUI: %s", err)
			continue
		}

		n.AppSKey = ns.AppSKey
		n.NwkSKey = ns.NwkSKey
		n.DevAddr = ns.DevAddr

		if err := storage.UpdateNode(ctx.DB, n); err != nil {
			log.WithField("dev_eui", n.DevEUI).Warningf("storage/nsmigrate: could not update node: %s", err)
			continue
		}

		log.WithField("dev_eui", n.DevEUI).Info("storage/nsmigrate: node-session data migrated")
	}
}
