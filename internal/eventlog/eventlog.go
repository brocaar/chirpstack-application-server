package eventlog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lorawan"
)

const (
	deviceEventUplinkPubSubKeyTempl = "lora:as:device:%s:pubsub:event"
)

// Event types.
const (
	Uplink   = "uplink"
	ACK      = "ack"
	Join     = "join"
	Error    = "error"
	Status   = "status"
	Location = "location"
)

// EventLog contains an event log.
type EventLog struct {
	Type    string
	Payload interface{}
}

// LogEventForDevice logs an event for the given device.
func LogEventForDevice(devEUI lorawan.EUI64, el EventLog) error {
	c := storage.RedisPool().Get()
	defer c.Close()

	key := fmt.Sprintf(deviceEventUplinkPubSubKeyTempl, devEUI)
	b, err := json.Marshal(el)
	if err != nil {
		return errors.Wrap(err, "gob encode error")
	}

	if _, err := c.Do("PUBLISH", key, b); err != nil {
		return errors.Wrap(err, "publish device event error")
	}

	return nil
}

// GetEventLogForDevice subscribes to the device events for the given DevEUI
// and sends this to the given channel.
func GetEventLogForDevice(ctx context.Context, devEUI lorawan.EUI64, eventsChan chan EventLog) error {
	c := storage.RedisPool().Get()
	defer c.Close()

	key := fmt.Sprintf(deviceEventUplinkPubSubKeyTempl, devEUI)
	psc := redis.PubSubConn{Conn: c}
	if err := psc.Subscribe(key); err != nil {
		return errors.Wrap(err, "subscribe error")
	}

	done := make(chan error, 1)

	go func() {
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				el, err := redisMessageToEventLog(v)
				if err != nil {
					log.WithError(err).Error("decode message errror")
				} else {
					eventsChan <- el
				}
			case redis.Subscription:
				if v.Count == 0 {
					done <- nil
					return
				}
			case error:
				done <- v
				return
			}
		}
	}()

	// todo: make this a config value?
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			if err := psc.Ping(""); err != nil {
				log.WithError(err).Error("subscription ping error")
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			return err
		}
	}

	if err := psc.Unsubscribe(); err != nil {
		return errors.Wrap(err, "unsubscribe error")
	}

	return <-done
}

func redisMessageToEventLog(msg redis.Message) (EventLog, error) {
	var el EventLog
	if err := json.Unmarshal(msg.Data, &el); err != nil {
		return el, errors.Wrap(err, "gob decode error")
	}

	return el, nil
}
