package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// CodeMigration checks if the given function code has been applied and if not
// it will execute the given function.
func CodeMigration(name string, f func(db sqlx.Ext) error) error {
	return Transaction(func(tx sqlx.Ext) error {
		_, err := tx.Exec("lock table code_migration")
		if err != nil {
			// The table might not exist when the code migration is executed
			// before the schema migrations.
			return ErrTransactionRollback
		}

		res, err := tx.Exec(`
			insert into code_migration (
				id,
				applied_at
			) values ($1, $2)
			on conflict
				do nothing
		`, name, time.Now())
		if err != nil {
			switch err := err.(type) {
			case *pq.Error:
				switch err.Code.Name() {
				case "unique_violation":
					return nil
				}
			}

			return err
		}

		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if ra == 0 {
			return nil
		}

		return f(tx)
	})
}
