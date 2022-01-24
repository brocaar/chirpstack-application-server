package loracloud

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
)

const (
	geolocBufferKeyTempl = "lora:as:device:%s:loracloud:rxinfo"
)

// SaveGeolocBuffer saves the given items in the geolocation buffer.
// It overwrites the previous buffer to make sure that expired items do not
// stay in the buffer as the TTL is set on the key, not on the items.
func SaveGeolocBuffer(ctx context.Context, devEUI lorawan.EUI64, items [][]*gw.UplinkRXInfo, ttl time.Duration) error {
	// nothing to do
	if ttl == 0 || len(items) == 0 {
		return nil
	}

	key := storage.GetRedisKey(geolocBufferKeyTempl, devEUI)
	pipe := storage.RedisClient().TxPipeline()
	pipe.Del(ctx, key)

	for i := range items {
		var frameRXInfo FrameRXInfo
		for j := range items[i] {
			frameRXInfo.RxInfo = append(frameRXInfo.RxInfo, items[i][j])
		}

		b, err := proto.Marshal(&frameRXInfo)
		if err != nil {
			return errors.Wrap(err, "protobuf marshal error")
		}

		pipe.RPush(ctx, key, b)
	}

	pipe.PExpire(ctx, key, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return errors.Wrap(err, "redis exec error")
	}

	return nil
}

// GetGeolocBuffer returns the geolocation buffer. Items that exceed the
// given TTL are not returned.
func GetGeolocBuffer(ctx context.Context, devEUI lorawan.EUI64, ttl time.Duration) ([][]*gw.UplinkRXInfo, error) {
	// nothing to do
	if ttl == 0 {
		return nil, nil
	}

	key := storage.GetRedisKey(geolocBufferKeyTempl, devEUI)
	resp, err := storage.RedisClient().LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, errors.Wrap(err, "read buffer error")
	}

	var out [][]*gw.UplinkRXInfo

	for _, b := range resp {
		var item FrameRXInfo
		if err := proto.Unmarshal([]byte(b), &item); err != nil {
			return nil, errors.Wrap(err, "protobuf unmarshal error")
		}

		add := true

		for _, rxInfo := range item.RxInfo {
			// skip frames without timestamp as we can't compare it against
			// the configured TTL
			if rxInfo.Time == nil {
				add = false
				break
			}

			ts, err := ptypes.Timestamp(rxInfo.Time)
			if err != nil {
				return nil, errors.Wrap(err, "get timestamp error")
			}

			// Ignore items before TTL as the TTL is set on the key of the buffer,
			// not on the item.
			if time.Now().Sub(ts) > ttl {
				add = false
				break
			}
		}

		if add {
			out = append(out, item.RxInfo)
		}
	}

	return out, nil
}
