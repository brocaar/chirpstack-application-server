package common

import (
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/loraserver/api/ns"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// DB holds the database connection pool.
var DB *sqlx.DB

// RedisPool holds the Redis connection pool.
var RedisPool *redis.Pool

// NetworkServer holds the connection to the network-server API.
var NetworkServer ns.NetworkServerClient

// Handler holds the handler of events.
var Handler handler.Handler
