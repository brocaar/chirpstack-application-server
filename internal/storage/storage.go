package storage

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/migrations"
)

var (
	jwtsecret      []byte
	HashIterations = 100000
)

// Setup configures the storage package.
func Setup(c config.Config) error {
	log.Info("storage: setting up storage package")

	jwtsecret = []byte(c.ApplicationServer.ExternalAPI.JWTSecret)
	HashIterations = c.General.PasswordHashIterations

	log.Info("storage: setting up Redis pool")
	redisPool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: c.Redis.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(c.Redis.URL,
				redis.DialReadTimeout(redisDialReadTimeout),
				redis.DialWriteTimeout(redisDialWriteTimeout),
			)
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Now().Sub(t) < onBorrowPingInterval {
				return nil
			}

			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}

	log.Info("storage: connecting to PostgreSQL database")
	d, err := sqlx.Open("postgres", c.PostgreSQL.DSN)
	if err != nil {
		return errors.Wrap(err, "storage: PostgreSQL connection error")
	}
	for {
		if err := d.Ping(); err != nil {
			log.WithError(err).Warning("storage: ping PostgreSQL database error, will retry in 2s")
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	db = &DBLogger{d}

	if c.PostgreSQL.Automigrate {
		log.Info("storage: applying PostgreSQL data migrations")
		m := &migrate.AssetMigrationSource{
			Asset:    migrations.Asset,
			AssetDir: migrations.AssetDir,
			Dir:      "",
		}
		n, err := migrate.Exec(db.DB.DB, "postgres", m, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "storage: applying PostgreSQL data migrations error")
		}
		log.WithField("count", n).Info("storage: PostgreSQL data migrations applied")
	}

	return nil
}
