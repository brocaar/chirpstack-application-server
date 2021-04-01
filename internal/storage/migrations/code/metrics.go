package code

import (
	"fmt"
	"time"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

// AggregationInterval defines the aggregation type.
type AggregationInterval string

// Metrics aggregation intervals.
const (
	AggregationMinute AggregationInterval = "MINUTE"
	AggregationHour   AggregationInterval = "HOUR"
	AggregationDay    AggregationInterval = "DAY"
	AggregationMonth  AggregationInterval = "MONTH"
)

const (
	metricsKeyTempl = "lora:as:metrics:{%s}:%s:%d" // metrics key (identifier | aggregation | timestamp)
)

var (
	timeLocation         = time.Local
	aggregationIntervals []AggregationInterval
	metricsMinuteTTL     time.Duration
	metricsHourTTL       time.Duration
	metricsDayTTL        time.Duration
	metricsMonthTTL      time.Duration
)

// MetricsRecord holds a single metrics record.
type MetricsRecord struct {
	Time    time.Time
	Metrics map[string]float64
}

// SaveMetrics stores the given metrics into Redis.
func SaveMetrics(conf config.Config, redisClient redis.UniversalClient, name string, metrics MetricsRecord) error {
	for _, agg := range aggregationIntervals {
		if err := SaveMetricsForInterval(conf, redisClient, agg, name, metrics); err != nil {
			return errors.Wrap(err, "save metrics for interval error")
		}
	}
	return nil
}

// SaveMetricsForInterval aggregates and stores the given metrics.
func SaveMetricsForInterval(conf config.Config, redisClient redis.UniversalClient, agg AggregationInterval, name string, metrics MetricsRecord) error {
	if len(metrics.Metrics) == 0 {
		return nil
	}

	var exp time.Duration
	ts := metrics.Time.In(timeLocation)

	// handle aggregation
	switch agg {
	case AggregationMinute:
		// truncate timestamp to minute precision
		ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), 0, 0, timeLocation)
		exp = metricsMinuteTTL
	case AggregationHour:
		// truncate timestamp to hour precision
		ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, timeLocation)
		exp = metricsHourTTL
	case AggregationDay:
		// truncate timestamp to day precision
		ts = time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, timeLocation)
		exp = metricsDayTTL
	case AggregationMonth:
		// truncate timestamp to month precision
		ts = time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, timeLocation)
		exp = metricsMonthTTL
	default:
		return fmt.Errorf("unexepcted aggregation interval: %s", agg)
	}

	key := GetRedisKey(conf.Redis.KeyPrefix, metricsKeyTempl, name, agg, ts.Unix())

	pipe := redisClient.TxPipeline()
	for k, v := range metrics.Metrics {
		pipe.HIncrByFloat(key, k, v)
	}
	pipe.PExpire(key, exp)
	if _, err := pipe.Exec(); err != nil {
		return errors.Wrap(err, "exec error")
	}
	return nil
}

// GetRedisKey returns the Redis key given a template and parameters.
func GetRedisKey(keyPrefix string, tmpl string, params ...interface{}) string {
	return keyPrefix + fmt.Sprintf(tmpl, params...)
}
