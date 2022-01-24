package storage

import (
	"database/sql"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	// register postgresql driver
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// Errors.
var (
	// ErrTransactionRollback indicates that the transaction must be rolled back,
	// but does not raise an error returned by the Transaction function.
	ErrTransactionRollback = errors.New("rollback")
)

// redisClient holds the Redis client.
var redisClient redis.UniversalClient

// db holds the PostgreSQL connection pool.
var db *DBLogger

// DBLogger is a DB wrapper which logs the executed sql queries and their
// duration.
type DBLogger struct {
	*sqlx.DB
}

// Beginx returns a transaction with logging.
func (db *DBLogger) Beginx() (*TxLogger, error) {
	tx, err := db.DB.Beginx()
	return &TxLogger{tx}, err
}

// Query logs the queries executed by the Query method.
func (db *DBLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.Query(query, args...)
	logQuery(query, time.Since(start), args...)
	return rows, err
}

// Queryx logs the queries executed by the Queryx method.
func (db *DBLogger) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := db.DB.Queryx(query, args...)
	logQuery(query, time.Since(start), args...)
	return rows, err
}

// QueryRowx logs the queries executed by the QueryRowx method.
func (db *DBLogger) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	row := db.DB.QueryRowx(query, args...)
	logQuery(query, time.Since(start), args...)
	return row
}

// Exec logs the queries executed by the Exec method.
func (db *DBLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := db.DB.Exec(query, args...)
	logQuery(query, time.Since(start), args...)
	return res, err
}

// TxLogger logs the executed sql queries and their duration.
type TxLogger struct {
	*sqlx.Tx
}

// Query logs the queries executed by the Query method.
func (q *TxLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := q.Tx.Query(query, args...)
	logQuery(query, time.Since(start), args...)
	return rows, err
}

// Queryx logs the queries executed by the Queryx method.
func (q *TxLogger) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := q.Tx.Queryx(query, args...)
	logQuery(query, time.Since(start), args...)
	return rows, err
}

// QueryRowx logs the queries executed by the QueryRowx method.
func (q *TxLogger) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	start := time.Now()
	row := q.Tx.QueryRowx(query, args...)
	logQuery(query, time.Since(start), args...)
	return row
}

// Exec logs the queries executed by the Exec method.
func (q *TxLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := q.Tx.Exec(query, args...)
	logQuery(query, time.Since(start), args...)
	return res, err
}

func logQuery(query string, duration time.Duration, args ...interface{}) {
	log.WithFields(log.Fields{
		"query":    query,
		"args":     args,
		"duration": duration,
	}).Debug("sql query executed")
}

// DB returns the PostgreSQL database object.
func DB() *DBLogger {
	return db
}

// RedisClient returns the RedisClient.
func RedisClient() redis.UniversalClient {
	return redisClient
}

// Transaction wraps the given function in a transaction. In case the given
// functions returns an error, the transaction will be rolled back.
func Transaction(f func(tx sqlx.Ext) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "storage: begin transaction error")
	}

	err = f(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Wrap(rbErr, "storage: transaction rollback error")
		}

		if err == ErrTransactionRollback {
			return nil
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "storage: transaction commit error")
	}
	return nil
}
