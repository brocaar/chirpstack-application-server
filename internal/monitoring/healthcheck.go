package monitoring

import (
	"net/http"

	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/pkg/errors"
)

func healthCheckHandlerFunc(w http.ResponseWriter, r *http.Request) {
	_, err := storage.RedisClient().Ping(r.Context()).Result()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(errors.Wrap(err, "redis ping error").Error()))
	}

	err = storage.DB().Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(errors.Wrap(err, "postgresql ping error").Error()))
	}

	w.WriteHeader(http.StatusOK)
}
