package eventlog

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lorawan"
)

const (
	deviceEventUplinkPubSubKeyTempl = "lora:as:device:%s:pubsub:event"
)

// Event types.
const (
	Uplink = "uplink"
	ACK    = "ack"
	Join   = "join"
	Error  = "error"
)

// EventLog contains an event log.
type EventLog struct {
	Type    string
	Payload interface{}
}

// LogEventForDevice logs an event for the given device.
func LogEventForDevice(devEUI lorawan.EUI64, el EventLog) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()

	key := fmt.Sprintf(deviceEventUplinkPubSubKeyTempl, devEUI)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(el); err != nil {
		return errors.Wrap(err, "gob encode error")
	}

	if _, err := c.Do("PUBLISH", key, buf.Bytes()); err != nil {
		return errors.Wrap(err, "publish device event error")
	}

	return nil
}

// GetEventLogForDevice subscribes to the device events for the given DevEUI
// and sends this to the given channel.
func GetEventLogForDevice(ctx context.Context, devEUI lorawan.EUI64, eventsChan chan EventLog) error {
	c := config.C.Redis.Pool.Get()
	defer c.Close()
	var done bool

	go func() {
		<-ctx.Done()
		done = true
		c.Close()
	}()

	key := fmt.Sprintf(deviceEventUplinkPubSubKeyTempl, devEUI)
	psc := redis.PubSubConn{Conn: c}
	if err := psc.Subscribe(key); err != nil {
		return errors.Wrap(err, "subscribe error")
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			el, err := redisMessageToEventLog(v)
			if err != nil {
				return errors.Wrap(err, "decode message error")
			}
			eventsChan <- el
		case error:
			if done {
				return nil
			}
			return errors.Wrap(v, "receive error")
		}
	}
}

func redisMessageToEventLog(msg redis.Message) (EventLog, error) {
	var el EventLog
	if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&el); err != nil {
		return el, errors.Wrap(err, "gob decode error")
	}

	return el, nil
}
