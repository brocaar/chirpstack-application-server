package code

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

// MigrateToClusterKeys migrates the keys to Redis Cluster compatible keys.
func MigrateToClusterKeys(conf config.Config) error {

	keys, err := storage.RedisClient().Keys(context.Background(), "lora:as:metrics:*").Result()
	if err != nil {
		return errors.Wrap(err, "get keys error")
	}

	for i, key := range keys {
		if err := migrateKey(conf, key); err != nil {
			log.WithError(err).Error("migrations/code: migrate metrics key error")
		}

		if i > 0 && i%1000 == 0 {
			log.WithFields(log.Fields{
				"migrated":    i,
				"total_count": len(keys),
			}).Info("migrations/code: migrating metrics keys")
		}
	}

	return nil
}

func migrateKey(conf config.Config, key string) error {
	keyParts := strings.Split(key, ":")
	if len(keyParts) < 6 {
		return fmt.Errorf("key %s is invalid", key)
	}

	ttlMap := map[string]time.Duration{
		"MINUTE": conf.Metrics.Redis.MinuteAggregationTTL,
		"HOUR":   conf.Metrics.Redis.HourAggregationTTL,
		"DAY":    conf.Metrics.Redis.DayAggregationTTL,
		"MONTH":  conf.Metrics.Redis.MonthAggregationTTL,
	}

	ttl, ok := ttlMap[keyParts[len(keyParts)-2]]
	if !ok {
		return fmt.Errorf("key %s is invalid", key)
	}

	newKey := storage.GetRedisKey("lora:as:metrics:{%s}:%s", strings.Join(keyParts[3:len(keyParts)-2], ":"), strings.Join(keyParts[len(keyParts)-2:], ":"))

	val, err := storage.RedisClient().HGetAll(context.Background(), key).Result()
	if err != nil {
		return errors.Wrap(err, "hgetall error")
	}

	pipe := storage.RedisClient().TxPipeline()
	for k, v := range val {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return errors.Wrap(err, "parse float error")
		}
		pipe.HIncrByFloat(context.Background(), newKey, k, f)
	}
	pipe.PExpire(context.Background(), key, ttl)

	if _, err := pipe.Exec(context.Background()); err != nil {
		return errors.Wrap(err, "exec error")
	}

	return nil
}
