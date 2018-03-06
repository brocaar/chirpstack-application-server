package storage

import (
	"fmt"
	"time"

	"github.com/gusseleet/lora-app-server/internal/common"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	// register postgresql driver
	_ "github.com/lib/pq"
)

const (
	redisMaxIdle        = 3
	redisIdleTimeoutSec = 240
)

// NewRedisPool returns a new Redis connection pool.
func NewRedisPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTimeoutSec * time.Second,
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

// OpenDatabase opens the database and performs a ping to make sure the
// database is up.
func OpenDatabase(dsn string) (*common.DBLogger, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %s", err)
	}
	for {
		if err := db.Ping(); err != nil {
			log.Errorf("ping database error, will retry in 2s: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	return &common.DBLogger{DB: db}, nil
}

// Transaction wraps the given function in a transaction. In case the given
// functions returns an error, the transaction will be rolled back.
func Transaction(db *common.DBLogger, f func(tx sqlx.Ext) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "begin transaction error")
	}

	err = f(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Wrap(rbErr, "transaction rollback error")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit error")
	}
	return nil
}
