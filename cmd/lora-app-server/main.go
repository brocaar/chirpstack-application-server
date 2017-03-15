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
	"time"

	log "github.com/Sirupsen/logrus"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/downlink"
	"github.com/brocaar/lora-app-server/internal/handler"
	"github.com/brocaar/lora-app-server/internal/migrations"
	"github.com/brocaar/lora-app-server/internal/static"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/lora-app-server/internal/storage/nsmigrate"
	"github.com/brocaar/loraserver/api/as"
	"github.com/brocaar/loraserver/api/ns"
)

func init() {
	grpclog.SetLogger(log.StandardLogger())
}

var version string // set by the compiler

func run(c *cli.Context) error {
	log.SetLevel(log.Level(uint8(c.Int("log-level"))))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://docs.loraserver.io/",
	}).Info("starting LoRa App Server")

	// get context
	lsCtx := mustGetContext(c)

	// migrate the database
	if c.Bool("db-automigrate") {
		log.Info("applying database migrations")
		m := &migrate.AssetMigrationSource{
			Asset:    migrations.Asset,
			AssetDir: migrations.AssetDir,
			Dir:      "",
		}
		n, err := migrate.Exec(lsCtx.DB.DB, "postgres", m, migrate.Up)
		if err != nil {
			log.Fatalf("applying migrations failed: %s", err)
		}
		log.WithField("count", n).Info("migrations applied")

	}

	if c.Bool("migrate-node-sessions") {
		log.Info("migrating node-session data from Redis")
		nsmigrate.Migrate(lsCtx)
	}

	// Set up the JWT secret for making tokens
	storage.SetUserSecret(c.String("jwt-secret"))
	// Set the password hash iterations
	storage.HashIterations = c.Int("pw-hash-iterations")

	// handle incoming downlink payloads
	go downlink.HandleDataDownPayloads(lsCtx, lsCtx.Handler.DataDownChan())

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

	// setup the client http interface variable
	// we need to start the gRPC service first, as it is used by the
	// grpc-gateway
	var clientHTTPHandler http.Handler

	// switch between gRPC and "plain" http handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			clientAPIHandler.ServeHTTP(w, r)
		} else {
			if clientHTTPHandler == nil {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}
			clientHTTPHandler.ServeHTTP(w, r)
		}
	})
	go func() {
		if c.String("http-tls-cert") == "" || c.String("http-tls-key") == "" {
			log.Fatal("--http-tls-cert (HTTP_TLS_CERT) and --http-tls-key (HTTP_TLS_KEY) must be set")
		}
		log.WithFields(log.Fields{
			"bind":     c.String("http-bind"),
			"tls-cert": c.String("http-tls-cert"),
			"tls-key":  c.String("http-tls-key"),
		}).Info("starting client api server")
		log.Fatal(http.ListenAndServeTLS(c.String("http-bind"), c.String("http-tls-cert"), c.String("http-tls-key"), handler))
	}()

	// give the http server some time to start
	time.Sleep(time.Millisecond * 100)

	// now the gRPC gateway has been started, attach the http handlers
	// (this will setup the grpc-gateway too)
	clientHTTPHandler = mustGetHTTPHandler(ctx, lsCtx, c)

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

	// setup redis pool
	log.Info("setup redis connection pool")
	rp := storage.NewRedisPool(c.String("redis-url"))

	// setup mqtt handler
	h, err := handler.NewMQTTHandler(rp, c.String("mqtt-server"), c.String("mqtt-username"), c.String("mqtt-password"))
	if err != nil {
		log.Fatalf("setup mqtt handler error: %s", err)
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
		RedisPool:     rp,
		NetworkServer: ns.NewNetworkServerClient(nsConn),
		Handler:       h,
	}
}

func mustGetClientAPIServer(ctx context.Context, lsCtx common.Context, c *cli.Context) *grpc.Server {
	var validator auth.Validator
	if c.String("jwt-secret") != "" {
		validator = auth.NewJWTValidator(lsCtx.DB, "HS256", c.String("jwt-secret"))
	} else {
		log.Fatal("--jwt-secret must be set")
	}

	gs := grpc.NewServer()
	pb.RegisterApplicationServer(gs, api.NewApplicationAPI(lsCtx, validator))
	pb.RegisterChannelListServer(gs, api.NewChannelListAPI(lsCtx, validator))
	pb.RegisterDownlinkQueueServer(gs, api.NewDownlinkQueueAPI(lsCtx, validator))
	pb.RegisterNodeServer(gs, api.NewNodeAPI(lsCtx, validator))
	pb.RegisterUserServer(gs, api.NewUserAPI(lsCtx, validator))
	pb.RegisterInternalServer(gs, api.NewInternalUserAPI(lsCtx, validator))

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

	if err := pb.RegisterApplicationHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register application handler error: %s", err)
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
	if err := pb.RegisterUserHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register user handler error: %s", err)
	}
	if err := pb.RegisterInternalHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		log.Fatalf("register internal handler error: %s", err)
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
		cli.BoolFlag{
			Name:   "migrate-node-sessions",
			Usage:  "migrate some of the node-session data to the application-server storage (run this once when migrating from loraserver 0.11.x)",
			EnvVar: "MIGRATE_NODE_SESSIONS",
		},
		cli.StringFlag{
			Name:   "redis-url",
			Usage:  "redis url (e.g. redis://user:password@hostname/0)",
			Value:  "redis://localhost:6379",
			EnvVar: "REDIS_URL",
		},
		cli.StringFlag{
			Name:   "mqtt-server",
			Usage:  "mqtt server (e.g. scheme://host:port where scheme is tcp, ssl or ws)",
			Value:  "tcp://localhost:1883",
			EnvVar: "MQTT_SERVER",
		},
		cli.StringFlag{
			Name:   "mqtt-username",
			Usage:  "mqtt server username (optional)",
			EnvVar: "MQTT_USERNAME",
		},
		cli.StringFlag{
			Name:   "mqtt-password",
			Usage:  "mqtt server password (optional)",
			EnvVar: "MQTT_PASSWORD",
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
			Usage:  "JWT secret used for api authentication / authorization",
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
		cli.IntFlag{
			Name:   "pw-hash-iterations",
			Usage:  "the number of iterations used to generate the password hash",
			Value:  100000,
			EnvVar: "PW_HASH_ITERATIONS",
		},
		cli.IntFlag{
			Name:   "log-level",
			Value:  4,
			Usage:  "debug=5, info=4, warning=3, error=2, fatal=1, panic=0",
			EnvVar: "LOG_LEVEL",
		},
	}
	app.Run(os.Args)
}
