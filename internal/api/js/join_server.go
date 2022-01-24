package js

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend/joinserver"
)

var (
	bind    string
	caCert  string
	tlsCert string
	tlsKey  string
)

// Setup configures the package.
func Setup(conf config.Config) error {
	bind = conf.JoinServer.Bind
	caCert = conf.JoinServer.CACert
	tlsCert = conf.JoinServer.TLSCert
	tlsKey = conf.JoinServer.TLSKey

	log.WithFields(log.Fields{
		"bind":     bind,
		"ca_cert":  caCert,
		"tls_cert": tlsCert,
		"tls_key":  tlsKey,
	}).Info("api/js: starting join-server api")

	handler, err := getHandler(conf)
	if err != nil {
		return errors.Wrap(err, "get join-server handler error")
	}

	server := http.Server{
		Handler:   handler,
		Addr:      bind,
		TLSConfig: &tls.Config{},
	}

	if caCert == "" && tlsCert == "" && tlsKey == "" {
		go func() {
			err := server.ListenAndServe()
			log.WithError(err).Fatal("join-server api error")
		}()
		return nil
	}

	if caCert != "" {
		caCert, err := ioutil.ReadFile(caCert)
		if err != nil {
			return errors.Wrap(err, "read ca certificate error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return errors.New("append ca certificate error")
		}

		server.TLSConfig.ClientCAs = caCertPool
		server.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert

		log.WithFields(log.Fields{
			"ca_cert": caCert,
		}).Info("api/js: join-server is configured with client-certificate authentication")
	}

	go func() {
		err := server.ListenAndServeTLS(tlsCert, tlsKey)
		log.WithError(err).Fatal("api/js: join-server api error")
	}()

	return nil
}
func getHandler(conf config.Config) (http.Handler, error) {
	jsConf := joinserver.HandlerConfig{
		Logger: log.StandardLogger(),
		GetDeviceKeysByDevEUIFunc: func(devEUI lorawan.EUI64) (joinserver.DeviceKeys, error) {
			dk, err := storage.GetDeviceKeys(context.TODO(), storage.DB(), devEUI)
			if err != nil {
				return joinserver.DeviceKeys{}, errors.Wrap(err, "get device-keys error")
			}

			if dk.JoinNonce == (1<<24)-1 {
				return joinserver.DeviceKeys{}, errors.New("join-nonce overflow")
			}
			dk.JoinNonce++
			if err := storage.UpdateDeviceKeys(context.TODO(), storage.DB(), &dk); err != nil {
				return joinserver.DeviceKeys{}, errors.Wrap(err, "update device-keys error")
			}

			return joinserver.DeviceKeys{
				DevEUI:    dk.DevEUI,
				NwkKey:    dk.NwkKey,
				AppKey:    dk.AppKey,
				JoinNonce: dk.JoinNonce,
			}, nil
		},
		GetKEKByLabelFunc: func(label string) ([]byte, error) {
			for _, kek := range conf.JoinServer.KEK.Set {
				if label == kek.Label {
					b, err := hex.DecodeString(kek.KEK)
					if err != nil {
						return nil, errors.Wrap(err, "decode hex encoded kek error")
					}

					return b, nil
				}
			}

			return nil, nil
		},
		GetASKEKLabelByDevEUIFunc: func(devEUI lorawan.EUI64) (string, error) {
			return conf.JoinServer.KEK.ASKEKLabel, nil
		},
		GetHomeNetIDByDevEUIFunc: func(devEUI lorawan.EUI64) (lorawan.NetID, error) {
			d, err := storage.GetDevice(context.TODO(), storage.DB(), devEUI, false, true)
			if err != nil {
				if errors.Cause(err) == storage.ErrDoesNotExist {
					return lorawan.NetID{}, joinserver.ErrDevEUINotFound
				}

				return lorawan.NetID{}, errors.Wrap(err, "get device error")
			}

			var netID lorawan.NetID

			netIDStr, ok := d.Variables.Map["home_netid"]
			if !ok {
				return netID, nil
			}

			if err := netID.UnmarshalText([]byte(netIDStr.String)); err != nil {
				return lorawan.NetID{}, errors.Wrap(err, "unmarshal netid error")
			}

			return netID, nil
		},
	}

	handler, err := joinserver.NewHandler(jsConf)
	if err != nil {
		return nil, errors.Wrap(err, "new join-server handler error")
	}

	return &prometheusMiddleware{
		handler:         handler,
		timingHistogram: conf.Metrics.Prometheus.APITimingHistogram,
	}, nil
}
