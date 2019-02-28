package test

import (
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/migrations"
)

func init() {
	config.C.ApplicationServer.ID = "6d5db27e-4ce2-4b2b-b5d7-91f069397978"
	config.C.ApplicationServer.API.PublicHost = "localhost:8001"
}

// GetConfig returns the test configuration.
func GetConfig() config.Config {
	log.SetLevel(log.ErrorLevel)

	var c config.Config

	c.PostgreSQL.DSN = "postgres://localhost/loraserver_as_test?sslmode=disable"
	c.Redis.URL = "redis://localhost:6379"
	c.ApplicationServer.Integration.MQTT.Server = "tcp://localhost:1883"

	if v := os.Getenv("TEST_POSTGRES_DSN"); v != "" {
		c.PostgreSQL.DSN = v
	}

	if v := os.Getenv("TEST_REDIS_URL"); v != "" {
		c.Redis.URL = v
	}

	if v := os.Getenv("TEST_MQTT_SERVER"); v != "" {
		c.ApplicationServer.Integration.MQTT.Server = v
	}

	if v := os.Getenv("TEST_MQTT_USERNAME"); v != "" {
		c.ApplicationServer.Integration.MQTT.Username = v
	}

	if v := os.Getenv("TEST_MQTT_PASSWORD"); v != "" {
		c.ApplicationServer.Integration.MQTT.Password = v
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
