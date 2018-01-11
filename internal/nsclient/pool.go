package nsclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/brocaar/loraserver/api/ns"
)

// Pool defines the network-server client pool.
type Pool interface {
	Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerClient, error)
}

type client struct {
	lastUsed time.Time
	client   ns.NetworkServerClient
}

type pool struct {
	sync.RWMutex
	clients map[string]client
}

// NewPool creates a Pool.
func NewPool() Pool {
	return &pool{
		clients: make(map[string]client),
	}
}

// Get returns a NetworkServerClient for the given server (hostname:ip).
func (p *pool) Get(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerClient, error) {
	defer p.Unlock()
	p.Lock()

	c, ok := p.clients[hostname]
	if !ok {
		nsClient, err := p.createClient(hostname, caCert, tlsCert, tlsKey)
		if err != nil {
			return nil, errors.Wrap(err, "create network-server api client error")
		}
		c = client{
			lastUsed: time.Now(),
			client:   nsClient,
		}
		p.clients[hostname] = c
	}

	return c.client, nil
}

func (p *pool) createClient(hostname string, caCert, tlsCert, tlsKey []byte) (ns.NetworkServerClient, error) {
	nsOpts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	if len(caCert) == 0 && len(tlsCert) == 0 && len(tlsKey) == 0 {
		nsOpts = append(nsOpts, grpc.WithInsecure())
		log.WithField("server", hostname).Warning("creating insecure network-server client")
	} else {
		log.WithField("server", hostname).Info("creating network-server client")
		cert, err := tls.X509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return nil, errors.Wrap(err, "load x509 keypair error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.Wrap(err, "append ca cert to pool error")
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
		return nil, errors.Wrap(err, "dial network-server api error")
	}

	return ns.NewNetworkServerClient(nsClient), nil
}
