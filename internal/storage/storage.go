package storage

import (
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	uuid "github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/migrations"
)

var (
	jwtsecret []byte
	// HashIterations denfines the number of times a password is hashed.
	HashIterations      = 100000
	applicationServerID uuid.UUID
)

// Setup configures the storage package.
func Setup(c config.Config) error {
	log.Info("storage: setting up storage package")

	jwtsecret = []byte(c.ApplicationServer.ExternalAPI.JWTSecret)
	HashIterations = c.General.PasswordHashIterations

	if err := applicationServerID.UnmarshalText([]byte(c.ApplicationServer.ID)); err != nil {
		return errors.Wrap(err, "decode application_server.id error")
	}

	log.Info("storage: setup metrics")
	// setup aggregation intervals
	var intervals []AggregationInterval
	for _, agg := range c.Metrics.Redis.AggregationIntervals {
		intervals = append(intervals, AggregationInterval(strings.ToUpper(agg)))
	}
	if err := SetAggregationIntervals(intervals); err != nil {
		return errors.Wrap(err, "set aggregation intervals error")
	}

	// setup timezone
	if err := SetTimeLocation(c.Metrics.Timezone); err != nil {
		return errors.Wrap(err, "set time location error")
	}

	// setup storage TTL
	SetMetricsTTL(
		c.Metrics.Redis.MinuteAggregationTTL,
		c.Metrics.Redis.HourAggregationTTL,
		c.Metrics.Redis.DayAggregationTTL,
		c.Metrics.Redis.MonthAggregationTTL,
	)

	log.Info("storage: setting up Redis client")

	opt := &redis.Options{}
	opt.PoolSize = c.Redis.PoolSize

	// URL config takes precedence
	if c.Redis.URL != "" {
		var err error
		opt, err = redis.ParseURL(c.Redis.URL)
		if err != nil {
			return errors.Wrap(err, "parse redis url error")
		}
	} else {
		opt.DB = c.Redis.Database
		opt.Password = c.Redis.Password
		opt.Addr = c.Redis.Servers[0]
	}

	if c.Redis.Cluster || c.Redis.MasterName != "" {
		var addresses []string
		if len(c.Redis.Servers) == 0 {
			addresses = []string{opt.Addr}
		} else {
			addresses = c.Redis.Servers
		}

		if c.Redis.Cluster {
			redisClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    addresses,
				PoolSize: opt.PoolSize,
				Password: opt.Password,
			})
		} else if c.Redis.MasterName != "" {
			redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:       c.Redis.MasterName,
				SentinelAddrs:    addresses,
				SentinelPassword: opt.Password,
				DB:               opt.DB,
				PoolSize:         opt.PoolSize,
				Password:         opt.Password,
			})
		}
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     opt.Addr,
			Password: opt.Password,
			PoolSize: opt.PoolSize,
			DB:       opt.DB,
		})
	}

	log.Info("storage: connecting to PostgreSQL database")
	d, err := sqlx.Open("postgres", c.PostgreSQL.DSN)
	if err != nil {
		return errors.Wrap(err, "storage: PostgreSQL connection error")
	}
	d.SetMaxOpenConns(c.PostgreSQL.MaxOpenConnections)
	d.SetMaxIdleConns(c.PostgreSQL.MaxIdleConnections)
	for {
		if err := d.Ping(); err != nil {
			log.WithError(err).Warning("storage: ping PostgreSQL database error, will retry in 2s")
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	db = &DBLogger{d}

	if c.PostgreSQL.Automigrate {
		log.Info("storage: applying PostgreSQL data migrations")
		m := &migrate.AssetMigrationSource{
			Asset:    migrations.Asset,
			AssetDir: migrations.AssetDir,
			Dir:      "",
		}
		n, err := migrate.Exec(db.DB.DB, "postgres", m, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "storage: applying PostgreSQL data migrations error")
		}
		log.WithField("count", n).Info("storage: PostgreSQL data migrations applied")
	}

	return nil
}
