package common

import (
	"time"

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

// GatewayPingFrequency holds the frequency used for sending gateway pings.
var GatewayPingFrequency int

// GatewayPingDR holds the data-rate used for sending gateway pings.
var GatewayPingDR int

// GatewayPingInterval holds the interval of the gateway ping.
var GatewayPingInterval time.Duration

// ApplicationServerID holds the application-server ID (UUID).
var ApplicationServerID string
