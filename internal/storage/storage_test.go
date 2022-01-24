package storage

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/chirpstack-application-server/internal/test"
)

// DatabaseTestSuiteBase provides the setup and teardown of the database
// for every test-run.
type DatabaseTestSuiteBase struct {
	suite.Suite
	tx *TxLogger
}

// SetupSuite is called once before starting the test-suite.
func (b *DatabaseTestSuiteBase) SetupSuite() {
	assert := require.New(b.T())
	conf := test.GetConfig()

	assert.NoError(Setup(conf))
	assert.NoError(MigrateDown(DB().DB))
	assert.NoError(MigrateUp(DB().DB))
}

// SetupTest is called before every test.
func (b *DatabaseTestSuiteBase) SetupTest() {
	tx, err := DB().Beginx()
	if err != nil {
		panic(err)
	}
	b.tx = tx

	if err := MigrateDown(DB().DB); err != nil {
		panic(err)
	}
	if err := MigrateUp(DB().DB); err != nil {
		panic(err)
	}

	RedisClient().FlushAll(context.Background())
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
	DatabaseTestSuiteBase
}

func TestStorage(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func TestGetRedisKey(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		keyPrefix string
		template  string
		params    []interface{}
		expected  string
	}{
		{
			keyPrefix: "as1:",
			template:  "foo:bar:key",
			expected:  "as1:foo:bar:key",
		},
		{
			template: "foo:bar:key",
			expected: "foo:bar:key",
		},
		{
			template: "foo:bar:%s",
			params:   []interface{}{"test"},
			expected: "foo:bar:test",
		},
	}

	for _, tst := range tests {
		keyPrefix = tst.keyPrefix
		out := GetRedisKey(tst.template, tst.params...)
		assert.Equal(tst.expected, out)
	}
}
