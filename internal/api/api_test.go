package api

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/test"
)

// DatabaseTestSuiteBase provides the setup and teardown of the database
// for every test-run.
type DatabaseTestSuiteBase struct {
	db *common.DBLogger
	tx *common.TxLogger
	p  *redis.Pool
}

// SetupSuite is called once before starting the test-suite.
func (b *DatabaseTestSuiteBase) SetupSuite() {
	conf := test.GetConfig()
	db, err := storage.OpenDatabase(conf.PostgresDSN)
	if err != nil {
		panic(err)
	}
	b.db = db

	b.p = storage.NewRedisPool(conf.RedisURL, 10, 0)

	config.C.PostgreSQL.DB = db
	config.C.Redis.Pool = b.p
}

// SetupTest is called before every test.
func (b *DatabaseTestSuiteBase) SetupTest() {
	tx, err := b.db.Beginx()
	if err != nil {
		panic(err)
	}
	b.tx = tx

	test.MustFlushRedis(b.p)
	test.MustResetDB(b.db)
}

// TearDownTest is called after every test.
func (b *DatabaseTestSuiteBase) TearDownTest() {
	if err := b.tx.Rollback(); err != nil {
		panic(err)
	}
}

// Tx returns a database transaction (which is rolled back after every
// test).
func (b *DatabaseTestSuiteBase) Tx() sqlx.Ext {
	return b.tx
}

// DB returns the database.
func (b *DatabaseTestSuiteBase) DB() *common.DBLogger {
	return b.db
}

// RedisPool returns the redis.Pool object.
func (b *DatabaseTestSuiteBase) RedisPool() *redis.Pool {
	return b.p
}

type APITestSuite struct {
	suite.Suite
	DatabaseTestSuiteBase
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
