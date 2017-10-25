package nsclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/brocaar/loraserver/api/ns"
	"github.com/pkg/errors"
)

// Pool defines the network-server client pool.
type Pool interface {
	Get(hostname string) (ns.NetworkServerClient, error)
}

type client struct {
	lastUsed time.Time
	client   ns.NetworkServerClient
}

type pool struct {
	sync.RWMutex
	caCert  string
	tlsCert string
	tlsKey  string
	clients map[string]client
}

// NewPool creates a Pool.
func NewPool(caCert, tlsCert, tlsKey string) Pool {
	return &pool{
		caCert:  caCert,
		tlsCert: tlsCert,
		tlsKey:  tlsKey,
		clients: make(map[string]client),
	}
}

// Get returns a NetworkServerClient for the given server (hostname:ip).
func (p *pool) Get(hostname string) (ns.NetworkServerClient, error) {
	defer p.Unlock()
	p.Lock()

	c, ok := p.clients[hostname]
	if !ok {
		nsClient, err := p.createClient(hostname)
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

func (p *pool) createClient(hostname string) (ns.NetworkServerClient, error) {
	nsOpts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	if p.tlsCert == "" && p.tlsKey == "" && p.caCert == "" {
		nsOpts = append(nsOpts, grpc.WithInsecure())
	} else {
		cert, err := tls.LoadX509KeyPair(p.tlsCert, p.tlsKey)
		if err != nil {
			return nil, errors.Wrap(err, "load x509 keypair error")
		}

		rawCACert, err := ioutil.ReadFile(p.caCert)
		if err != nil {
			return nil, errors.Wrap(err, "load ca cert error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(rawCACert) {
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
