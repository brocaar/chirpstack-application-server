package test

import (
	"os"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/migrations"
)

func init() {
	config.C.ApplicationServer.ID = "6d5db27e-4ce2-4b2b-b5d7-91f069397978"
	config.C.ApplicationServer.API.PublicHost = "localhost:8001"
}

// GetConfig returns the test configuration.
func GetConfig() config.Config {
	log.SetLevel(log.ErrorLevel)

	var c config.Config

	c.PostgreSQL.DSN = "postgres://localhost/chirpstack_as_test?sslmode=disable"
	c.Redis.URL = "redis://localhost:6379"
	c.ApplicationServer.Integration.MQTT.Server = "tcp://localhost:1883"
	c.ApplicationServer.ID = "6d5db27e-4ce2-4b2b-b5d7-91f069397978"
	c.ApplicationServer.Integration.AMQP.EventRoutingKeyTemplate = "application.{{ .ApplicationID }}.device.{{ .DevEUI }}.event.{{ .EventType }}"
	c.ApplicationServer.Integration.Kafka.Topic = "chirpstack_as"

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

	if v := os.Getenv("TEST_RABBITMQ_URL"); v != "" {
		c.ApplicationServer.Integration.AMQP.URL = v
	}

	if v := os.Getenv("TEST_KAFKA_BROKER"); v != "" {
		c.ApplicationServer.Integration.Kafka.Brokers = []string{v}
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
