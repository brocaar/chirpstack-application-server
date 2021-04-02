package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func (ts *StorageTestSuite) TestCodeMigration() {
	assert := require.New(ts.T())

	count := 0

	// returning an error does not mark the migration as completed
	assert.Error(CodeMigration("test_1", func(db sqlx.Ext) error {
		count++
		return fmt.Errorf("BOOM")
	}))

	assert.Equal(1, count)

	// re-run the migration
	assert.NoError(CodeMigration("test_1", func(db sqlx.Ext) error {
		count++
		return nil
	}))

	assert.Equal(2, count)

	// the migration has already been completed
	assert.NoError(CodeMigration("test_1", func(db sqlx.Ext) error {
		count++
		return nil
	}))

	assert.Equal(2, count)

	// new migration should run
	assert.NoError(CodeMigration("test_2", func(db sqlx.Ext) error {
		count++
		return nil
	}))

	assert.Equal(3, count)

	// migration has already been applied
	assert.NoError(CodeMigration("test_2", func(db sqlx.Ext) error {
		count++
		return nil
	}))

	assert.Equal(3, count)
}
