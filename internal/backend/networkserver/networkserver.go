package networkserver

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"sync"
	"time"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"

	"github.com/brocaar/chirpstack-api/go/v3/ns"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/logging"
)

var p Pool

// Pool defines the network-server client pool.
type Pool interface {
	Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerServiceClient, error)
}

type client struct {
	client     ns.NetworkServerServiceClient
	clientConn *grpc.ClientConn
	caCert     []byte
	tlsCert    []byte
	tlsKey     []byte
}

// Setup configures the networkserver package.
func Setup(conf config.Config) error {
	p = &pool{
		clients: make(map[string]client),
	}
	return nil
}

// GetPool returns the networkserver pool.
func GetPool() Pool {
	return p
}

// SetPool sets the network-server pool.
func SetPool(pp Pool) {
	p = pp
}

type pool struct {
	sync.RWMutex
	clients map[string]client
}

// Get returns a NetworkServerClient for the given server (hostname:ip).
func (p *pool) Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerServiceClient, error) {
	defer p.Unlock()
	p.Lock()

	var connect bool
	c, ok := p.clients[hostname]
	if !ok {
		connect = true
	}

	// if the connection exists in the map, but when the certificates changed
	// try to cloe the connection and re-connect
	if ok && (!bytes.Equal(c.caCert, caCert) || !bytes.Equal(c.tlsCert, tlsCert) || !bytes.Equal(c.tlsKey, tlsKey)) {
		c.clientConn.Close()
		delete(p.clients, hostname)
		connect = true
	}

	if connect {
		clientConn, nsClient, err := p.createClient(hostname, caCert, tlsCert, tlsKey)
		if err != nil {
			return nil, errors.Wrap(err, "create network-server api client error")
		}
		c = client{
			client:     nsClient,
			clientConn: clientConn,
			caCert:     caCert,
			tlsCert:    tlsCert,
			tlsKey:     tlsKey,
		}
		p.clients[hostname] = c
	}

	return c.client, nil
}

func (p *pool) createClient(hostname string, caCert, tlsCert, tlsKey []byte) (*grpc.ClientConn, ns.NetworkServerServiceClient, error) {
	logrusEntry := log.NewEntry(log.StandardLogger())
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	nsOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			logging.UnaryClientCtxIDInterceptor,
		),
		grpc.WithStreamInterceptor(
			grpc_logrus.StreamClientInterceptor(logrusEntry, logrusOpts...),
		),
		grpc.WithBalancerName(roundrobin.Name),
	}

	if len(caCert) == 0 && len(tlsCert) == 0 && len(tlsKey) == 0 {
		nsOpts = append(nsOpts, grpc.WithInsecure())
		log.WithField("server", hostname).Warning("creating insecure network-server client")
	} else {
		log.WithField("server", hostname).Info("creating network-server client")
		cert, err := tls.X509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "load x509 keypair error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, nil, errors.New("append ca cert to pool error")
		}

		nsOpts = append(nsOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		})))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	nsClient, err := grpc.DialContext(ctx, hostname, nsOpts...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "dial network-server api error")
	}

	return nsClient, ns.NewNetworkServerServiceClient(nsClient), nil
}
