package common

import "github.com/jmoiron/sqlx"

type Context struct {
	DB *sqlx.DB
}
