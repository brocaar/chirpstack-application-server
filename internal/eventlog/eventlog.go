package eventlog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

const (
	deviceEventStreamKey = "lora:as:device:%s:stream:event"
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

// PublishedAtMessage implements the proto.Message with PublishedAt getter.
type PublishedAtMessage interface {
	proto.Message
	GetPublishedAt() *timestamp.Timestamp
}

// EventLog contains an event log.
type EventLog struct {
	Type        string
	PublishedAt *timestamp.Timestamp
	Payload     json.RawMessage
	StreamID    string
}

// LogEventForDevice logs an event for the given device.
func LogEventForDevice(devEUI lorawan.EUI64, t string, msg proto.Message) error {
	conf := config.Get()

	if conf.Monitoring.PerDeviceEventLogMaxHistory > 0 {
		b, err := proto.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "marshal event error")
		}

		key := storage.GetRedisKey(deviceEventStreamKey, devEUI)
		pipe := storage.RedisClient().TxPipeline()

		pipe.XAdd(context.Background(), &redis.XAddArgs{
			Stream: key,
			MaxLen: conf.Monitoring.PerDeviceEventLogMaxHistory,
			Values: map[string]interface{}{
				"event": t,
				"data":  b,
			},
		})
		pipe.Expire(context.Background(), key, time.Hour*24*31)

		_, err = pipe.Exec(context.Background())
		if err != nil {
			return errors.Wrap(err, "redis xadd error")
		}
	}

	return nil
}

// GetEventLogForDevice subscribes to the device events for the given DevEUI
// and sends this to the given channel.
func GetEventLogForDevice(ctx context.Context, devEUI lorawan.EUI64, eventsChan chan EventLog) error {
	key := storage.GetRedisKey(deviceEventStreamKey, devEUI)
	lastID := "0"

	for {
		resp, err := storage.RedisClient().XRead(ctx, &redis.XReadArgs{
			Streams: []string{key, lastID},
			Count:   10,
			Block:   0,
		}).Result()
		if err != nil {
			if err == context.Canceled {
				return nil
			}

			return errors.Wrap(err, "redis stream error")
		}

		if len(resp) != 1 {
			return errors.New("exactly one stream response expected")
		}

		for _, msg := range resp[0].Messages {
			lastID = msg.ID
			var pl PublishedAtMessage

			event, ok := msg.Values["event"].(string)
			if !ok {
				continue
			}

			switch event {
			case Uplink:
				pl = &integration.UplinkEvent{}
			case ACK:
				pl = &integration.AckEvent{}
			case Join:
				pl = &integration.JoinEvent{}
			case Error:
				pl = &integration.ErrorEvent{}
			case Status:
				pl = &integration.StatusEvent{}
			case Location:
				pl = &integration.LocationEvent{}
			case TxAck:
				pl = &integration.TxAckEvent{}
			case Integration:
				pl = &integration.IntegrationEvent{}
			default:
				continue
			}

			b, ok := msg.Values["data"].(string)
			if !ok {
				continue
			}

			// decode the binary data
			if err := proto.Unmarshal([]byte(b), pl); err != nil {
				return errors.Wrap(err, "unmarshal protobuf error")
			}

			// encode it to json for display purposes
			jsonB, err := marshaler.Marshal(marshaler.ProtobufJSON, pl)
			if err != nil {
				return errors.Wrap(err, "marshal protobuf json error")
			}

			eventsChan <- EventLog{
				Type:        event,
				Payload:     json.RawMessage(jsonB),
				PublishedAt: pl.GetPublishedAt(),
				StreamID:    msg.ID,
			}
		}
	}

}
