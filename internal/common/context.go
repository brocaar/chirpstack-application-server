package common

import (
	"github.com/brocaar/loraserver/api/ns"
	"github.com/jmoiron/sqlx"
)

type Context struct {
	DB            *sqlx.DB
	NetworkServer ns.NetworkServerClient
}
