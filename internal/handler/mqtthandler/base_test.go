package mqtthandler

import (
	"fmt"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
)

type config struct {
	PostgresDSN  string
	RedisURL     string
	MQTTServer   string
	MQTTUsername string
	MQTTPassword string
}

// GetConfig returns the test configuration.
func GetConfig() *config {
	log.SetLevel(log.ErrorLevel)

	c := &config{
		RedisURL:   "redis://localhost:6379/15",
		MQTTServer: "tcp://localhost:1883",
	}

	if v := os.Getenv("TEST_REDIS_URL"); v != "" {
		c.RedisURL = v
	}

	if v := os.Getenv("TEST_MQTT_SERVER"); v != "" {
		c.MQTTServer = v
	}

	if v := os.Getenv("TEST_MQTT_USERNAME"); v != "" {
		c.MQTTUsername = v
	}

	if v := os.Getenv("TEST_MQTT_PASSWORD"); v != "" {
		c.MQTTPassword = v
	}

	return c
}

// MustFlushRedis flushes the Redis storage.
func MustFlushRedis(p *redis.Pool) {
	c := p.Get()
	defer c.Close()
	if _, err := c.Do("FLUSHALL"); err != nil {
		log.Fatal(err)
	}
}

// NewRedisPool returns a new Redis connection pool.
func NewRedisPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}
}
