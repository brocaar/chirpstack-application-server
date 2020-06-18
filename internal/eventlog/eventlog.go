package eventlog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v7"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

const (
	deviceEventUplinkPubSubKeyTempl = "lora:as:device:%s:pubsub:event"
)

// Event types.
const (
	Uplink      = "up"
	ACK         = "ack"
	Join        = "join"
	Error       = "error"
	Status      = "status"
	Location    = "location"
	TxAck       = "txack"
	Integration = "integration"
)

// EventLog contains an event log.
type EventLog struct {
	Type    string
	Payload json.RawMessage
}

// LogEventForDevice logs an event for the given device.
func LogEventForDevice(devEUI lorawan.EUI64, t string, msg proto.Message) error {
	b, err := marshaler.Marshal(marshaler.ProtobufJSON, msg)
	if err != nil {
		return errors.Wrap(err, "marshal protobuf json error")
	}

	el := EventLog{
		Type:    t,
		Payload: json.RawMessage(b),
	}

	key := fmt.Sprintf(deviceEventUplinkPubSubKeyTempl, devEUI)
	b, err = json.Marshal(el)
	if err != nil {
		return errors.Wrap(err, "json encode error")
	}

	err = storage.RedisClient().Publish(key, b).Err()
	if err != nil {
		return errors.Wrap(err, "publish device event error")
	}

	return nil
}

// GetEventLogForDevice subscribes to the device events for the given DevEUI
// and sends this to the given channel.
func GetEventLogForDevice(ctx context.Context, devEUI lorawan.EUI64, eventsChan chan EventLog) error {
	key := fmt.Sprintf(deviceEventUplinkPubSubKeyTempl, devEUI)

	sub := storage.RedisClient().Subscribe(key)
	_, err := sub.Receive()
	if err != nil {
		return errors.Wrap(err, "subscribe error")
	}

	ch := sub.Channel()

	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			el, err := redisMessageToEventLog(msg)
			if err != nil {
				log.WithError(err).Error("decode message error")
			} else {
				eventsChan <- el
			}
		case <-ctx.Done():
			sub.Close()
			return nil
		}
	}
}

func redisMessageToEventLog(msg *redis.Message) (EventLog, error) {
	var el EventLog
	if err := json.Unmarshal([]byte(msg.Payload), &el); err != nil {
		return el, errors.Wrap(err, "unmarshal message error")
	}

	return el, nil
}
