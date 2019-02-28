package storage

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lora-app-server/internal/test"
)

// DatabaseTestSuiteBase provides the setup and teardown of the database
// for every test-run.
type DatabaseTestSuiteBase struct {
	tx *TxLogger
}

// SetupSuite is called once before starting the test-suite.
func (b *DatabaseTestSuiteBase) SetupSuite() {
	conf := test.GetConfig()
	if err := Setup(conf); err != nil {
		panic(err)
	}
}

// SetupTest is called before every test.
func (b *DatabaseTestSuiteBase) SetupTest() {
	tx, err := DB().Beginx()
	if err != nil {
		panic(err)
	}
	b.tx = tx

	test.MustResetDB(DB().DB)
	test.MustFlushRedis(RedisPool())
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

type StorageTestSuite struct {
	suite.Suite
	DatabaseTestSuiteBase
}

func TestStorage(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
