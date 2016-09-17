package common

import (
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
)

type Context struct {
	DB            *sqlx.DB
	RedisPool     *redis.Pool
	NetworkServer ns.NetworkServerClient
	Handler       handler.Handler
}
