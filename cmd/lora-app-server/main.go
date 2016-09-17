//go:generate go-bindata -prefix ../../migrations/ -pkg migrations -o ../../internal/migrations/migrations_gen.go ../../migrations/
//go:generate go-bindata -prefix ../../static/ -pkg static -o ../../internal/static/static_gen.go ../../static/...

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/static"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/ns"
)

func init() {
	grpclog.SetLogger(log.StandardLogger())
}

var version string // set by the compiler

func run(c *cli.Context) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://docs.loraserver.io/",
	}).Info("starting LoRa App Server")

	// get context
	lsCtx := mustGetContext(c)

	// start the application-server api
	log.WithFields(log.Fields{
		"bind":     c.String("bind"),
		"ca-cert":  c.String("ca-cert"),
		"tls-cert": c.String("tls-cert"),
		"tls-key":  c.String("tls-key"),
	}).Info("starting application-server api")
	apiServer := mustGetAPIServer(lsCtx, c)
	ln, err := net.Listen("tcp", c.String("bind"))
	if err != nil {
		log.Fatalf("start application-server api listener error: %s", err)
	}
	go apiServer.Serve(ln)

	// setup the client api interface
	clientAPIHandler := mustGetClientAPIServer(ctx, lsCtx, c)

	// setup the client http interface
	clientHTTPHandler := mustGetHTTPHandler(ctx, lsCtx, c)

	// switch between gRPC and "plain" http handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			clientAPIHandler.ServeHTTP(w, r)
		} else {
			clientHTTPHandler.ServeHTTP(w, r)
		}
	})
	go func() {
		log.WithFields(log.Fields{
			"bind":     c.String("http-bind"),
			"tls-cert": c.String("http-tls-cert"),
			"tls-key":  c.String("http-tls-key"),
		}).Info("starting client api server")
		log.Fatal(http.ListenAndServeTLS(c.String("http-bind"), c.String("http-tls-cert"), c.String("http-tls-key"), handler))
	}()

	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping lora-app-server")
		// todo: handle graceful shutdown?
		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stopping immediately")
	}

	return nil
}

func mustGetContext(c *cli.Context) common.Context {
	log.Info("connecting to postgresql")
	db, err := storage.OpenDatabase(c.String("postgres-dsn"))
	if err != nil {
		log.Fatalf("database connection error: %s", err)
	}

	// setup network-server client
	log.WithFields(log.Fields{
		"server":   c.String("ns-server"),
		"ca-cert":  c.String("ns-ca-cert"),
		"tls-cert": c.String("ns-tls-cert"),
		"tls-key":  c.String("ns-tls-key"),
	}).Info("connecting to network-server api")
	var nsOpts []grpc.DialOption
	if c.String("ns-tls-cert") != "" && c.String("ns-tls-key") != "" {
		nsOpts = append(nsOpts, grpc.WithTransportCredentials(
			mustGetTransportCredentials(c.String("ns-tls-cert"), c.String("ns-tls-key"), c.String("ns-ca-cert"), false),
		))
	} else {
		nsOpts = append(nsOpts, grpc.WithInsecure())
	}

	nsConn, err := grpc.Dial(c.String("ns-server"), nsOpts...)
	if err != nil {
		log.Fatalf("network-server dial error: %s", err)
	}

	return common.Context{
		DB:            db,
		NetworkServer: ns.NewNetworkServerClient(nsConn),
	}
}

func mustGetClientAPIServer(ctx context.Context, lsCtx common.Context, c *cli.Context) *grpc.Server {
	var validator auth.Validator
	if c.String("jwt-secret") != "" {
		validator = auth.NewJWTValidator("HS256", c.String("jwt-secret"))
	} else {
		log.Warning("client api authentication and authorization is disabled (set jwt-secret to enable)")
		validator = auth.NopValidator{}
	}

	gs := grpc.NewServer()
	pb.RegisterChannelServer(gs, api.NewChannelAPI(lsCtx, validator))
	pb.RegisterChannelListServer(gs, api.NewChannelListAPI(lsCtx, validator))
	pb.RegisterDownlinkQueueServer(gs, api.NewDownlinkQueueAPI(lsCtx, validator))
	pb.RegisterNodeServer(gs, api.NewNodeAPI(lsCtx, validator))
	//pb.RegisterNodeSessionServer(gs, api.NewNodeSessionAPI(lsCtx, validator))

	return gs
}

func mustGetAPIServer(ctx common.Context, c *cli.Context) *grpc.Server {
	var opts []grpc.ServerOption
	if c.String("tls-cert") != "" && c.String("tls-key") != "" {
		creds := mustGetTransportCredentials(c.String("tls-cert"), c.String("tls-key"), c.String("ca-cert"), false)
		opts = append(opts, grpc.Creds(creds))
	}
	gs := grpc.NewServer(opts...)
	asAPI := api.NewApplicationServerAPI(ctx)
	as.RegisterApplicationServerServer(gs, asAPI)
	return gs
}

func mustGetHTTPHandler(ctx context.Context, lsCtx common.Context, c *cli.Context) http.Handler {
	r := mux.NewRouter()

	// setup json api handler
	jsonHandler := mustGetJSONGateway(ctx, lsCtx, c)
	log.WithField("path", "/api").Info("registering rest api handler and documentation endpoint")
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		data, err := static.Asset("swagger/index.html")
		if err != nil {
			log.Errorf("get swagger template error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}).Methods("get")
	r.PathPrefix("/api").Handler(jsonHandler)

	// setup static file server
	r.PathPrefix("/").Handler(http.FileServer(&assetfs.AssetFS{
		Asset:     static.Asset,
		AssetDir:  static.AssetDir,
		AssetInfo: static.AssetInfo,
		Prefix:    "",
	}))

	return r
}

func mustGetJSONGateway(ctx context.Context, lsCtx common.Context, c *cli.Context) http.Handler {
	// dial options for the grpc-gateway
	b, err := ioutil.ReadFile(c.String("http-tls-cert"))
	if err != nil {
		log.Fatalf("read http-tls-cert cert error: %s", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		log.Fatal("failed to append certificate")
	}
	grpcDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		// given the grpc-gateway is always connecting to localhost, does
		// InsecureSkipVerify=true cause any security issues?
		InsecureSkipVerify: true,
		RootCAs:            cp,
	}))}

	bindParts := strings.SplitN(c.String("http-bind"), ":", 2)
	if len(bindParts) != 2 {
		log.Fatal("get port from bind failed")
	}
	apiEndpoint := fmt.Sprintf("localhost:%s", bindParts[1])

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{
			EnumsAsInts:  false,
			EmitDefaults: true,
		},
	))

	if err := pb.RegisterChannelHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register channel handler error: %s", err)
	}
	if err := pb.RegisterChannelListHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register channel-list handler error: %s", err)
	}
	if err := pb.RegisterDownlinkQueueHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register downlink queue handler error: %s", err)
	}
	if err := pb.RegisterNodeHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register node handler error: %s", err)
	}
	if err := pb.RegisterNodeSessionHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register node-session handler error: %s", err)
	}

	return mux
}

func mustGetTransportCredentials(tlsCert, tlsKey, caCert string, verifyClientCert bool) credentials.TransportCredentials {
	var caCertPool *x509.CertPool
	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		log.WithFields(log.Fields{
			"cert": tlsCert,
			"key":  tlsKey,
		}).Fatalf("load key-pair error: %s", err)
	}

	if caCert != "" {
		rawCaCert, err := ioutil.ReadFile(caCert)
		if err != nil {
			log.WithField("ca", caCert).Fatalf("load ca cert error: %s", err)
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(rawCaCert)
	}

	if verifyClientCert {
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})
	} else {
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		})
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "lora-app-server"
	app.Usage = "application-server for LoRaWAN networks"
	app.Version = version
	app.Copyright = "See http://github.com/brocaar/lora-app-server for copyright information"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "postgres-dsn",
			Usage:  "postgresql dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable)",
			Value:  "postgres://localhost/loraserver?sslmode=disable",
			EnvVar: "POSTGRES_DSN",
		},
		cli.BoolFlag{
			Name:   "db-automigrate",
			Usage:  "automatically apply database migrations",
			EnvVar: "DB_AUTOMIGRATE",
		},
		cli.StringFlag{
			Name:   "ca-cert",
			Usage:  "ca certificate used by the api server (optional)",
			EnvVar: "CA_CERT",
		},
		cli.StringFlag{
			Name:   "tls-cert",
			Usage:  "tls certificate used by the api server (optional)",
			EnvVar: "TLS_CERT",
		},
		cli.StringFlag{
			Name:   "tls-key",
			Usage:  "tls key used by the api server (optional)",
			EnvVar: "TLS_KEY",
		},
		cli.StringFlag{
			Name:   "bind",
			Usage:  "ip:port to bind the api server",
			Value:  "0.0.0.0:8001",
			EnvVar: "BIND",
		},
		cli.StringFlag{
			Name:   "http-bind",
			Usage:  "ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api)",
			Value:  "0.0.0.0:8080",
			EnvVar: "HTTP_BIND",
		},
		cli.StringFlag{
			Name:   "http-tls-cert",
			Usage:  "http server TLS certificate",
			EnvVar: "HTTP_TLS_CERT",
		},
		cli.StringFlag{
			Name:   "http-tls-key",
			Usage:  "http server TLS key",
			EnvVar: "HTTP_TLS_KEY",
		},
		cli.StringFlag{
			Name:   "jwt-secret",
			Usage:  "JWT secret used for api authentication / authorization (disabled when left blank)",
			EnvVar: "JWT_SECRET",
		},
		cli.StringFlag{
			Name:   "ns-server",
			Usage:  "hostname:port of the network-server api server",
			Value:  "127.0.0.1:8000",
			EnvVar: "NS_SERVER",
		},
		cli.StringFlag{
			Name:   "ns-ca-cert",
			Usage:  "ca certificate used by the network-server client (optional)",
			EnvVar: "NS_CA_CERT",
		},
		cli.StringFlag{
			Name:   "ns-tls-cert",
			Usage:  "tls certificate used by the network-server client (optional)",
			EnvVar: "NS_TLS_CERT",
		},
		cli.StringFlag{
			Name:   "ns-tls-key",
			Usage:  "tls key used by the network-server client (optional)",
			EnvVar: "NS_TLS_KEY",
		},
	}
	app.Run(os.Args)
}
