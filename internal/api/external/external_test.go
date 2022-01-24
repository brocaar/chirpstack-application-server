package external

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/chirpstack-application-server/internal/test"
)

// DatabaseTestSuiteBase provides the setup and teardown of the database
// for every test-run.
type DatabaseTestSuiteBase struct {
	suite.Suite
	tx *storage.TxLogger
}

// SetupSuite is called once before starting the test-suite.
func (b *DatabaseTestSuiteBase) SetupSuite() {
	conf := test.GetConfig()
	conf.Monitoring.PerDeviceEventLogMaxHistory = 10
	config.Set(conf)

	if err := storage.Setup(conf); err != nil {
		panic(err)
	}
}

// SetupTest is called before every test.
func (b *DatabaseTestSuiteBase) SetupTest() {
	assert := require.New(b.T())

	tx, err := storage.DB().Beginx()
	if err != nil {
		panic(err)
	}
	b.tx = tx

	storage.RedisClient().FlushAll(context.Background())
	assert.NoError(storage.MigrateDown(storage.DB().DB))
	assert.NoError(storage.MigrateUp(storage.DB().DB))
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

type APITestSuite struct {
	DatabaseTestSuiteBase
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
