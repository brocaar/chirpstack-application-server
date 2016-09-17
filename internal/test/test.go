package test

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/migrations"
	"github.com/brocaar/lorawan"
)

// Config contains the test configuration.
type Config struct {
	PostgresDSN  string
	RedisURL     string
	MQTTServer   string
	MQTTUsername string
	MQTTPassword string
}

// GetConfig returns the test configuration.
func GetConfig() *Config {
	log.SetLevel(log.ErrorLevel)

	c := &Config{}

	if v := os.Getenv("TEST_POSTGRES_DSN"); v != "" {
		c.PostgresDSN = v
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

// MustResetDB re-applies all database migrations.
func MustResetDB(db *sqlx.DB) {
	m := &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "",
	}
	if _, err := migrate.Exec(db.DB, "postgres", m, migrate.Down); err != nil {
		log.Fatal(err)
	}
	if _, err := migrate.Exec(db.DB, "postgres", m, migrate.Up); err != nil {
		log.Fatal(err)
	}
}

// MustFlushRedis flushes the Redis storage.
func MustFlushRedis(p *redis.Pool) {
	c := p.Get()
	defer c.Close()
	if _, err := c.Do("FLUSHALL"); err != nil {
		log.Fatal(err)
	}
}

// Handler implements a Handler for testing.
type Handler struct {
	SendDataUpChan            chan handler.DataUpPayload
	SendJoinNotificationChan  chan handler.JoinNotification
	SendACKNotificationChan   chan handler.ACKNotification
	SendErrorNotificationChan chan handler.ErrorNotification
	DataDownPayloadChan       chan handler.DataDownPayload
}

func NewHandler() *Handler {
	return &Handler{
		SendDataUpChan:            make(chan handler.DataUpPayload, 100),
		SendJoinNotificationChan:  make(chan handler.JoinNotification, 100),
		SendACKNotificationChan:   make(chan handler.ACKNotification, 100),
		SendErrorNotificationChan: make(chan handler.ErrorNotification, 100),
		DataDownPayloadChan:       make(chan handler.DataDownPayload, 100),
	}
}

func (t *Handler) Close() error {
	return nil
}

func (t *Handler) SendDataUp(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload handler.DataUpPayload) error {
	t.SendDataUpChan <- payload
	return nil
}

func (t *Handler) SendJoinNotification(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload handler.JoinNotification) error {
	t.SendJoinNotificationChan <- payload
	return nil
}

func (t *Handler) SendACKNotification(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload handler.ACKNotification) error {
	t.SendACKNotificationChan <- payload
	return nil
}

func (t *Handler) SendErrorNotification(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload handler.ErrorNotification) error {
	t.SendErrorNotificationChan <- payload
	return nil
}

func (t *Handler) DataDownChan() chan handler.DataDownPayload {
	return t.DataDownPayloadChan
}
