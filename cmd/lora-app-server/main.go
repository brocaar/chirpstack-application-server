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
	"github.com/pkg/errors"
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
	"github.com/brocaar/lora-app-server/internal/storage/gwmigrate"
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

	tasks := []func(*cli.Context) error{
		setLogLevel,
		printStartMessage,
		setPostgreSQLConnection,
		setRedisPool,
		setHandler,
		setNetworkServerClient,
		runDatabaseMigrations,
		setJWTSecret,
		setHashIterations,
		setDisableAssignExistingUsers,
		handleDataDownPayloads,
		startApplicationServerAPI,
		startClientAPI(ctx),
	}

	for _, t := range tasks {
		if err := t(c); err != nil {
			log.Fatal(err)
		}
	}

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

func setLogLevel(c *cli.Context) error {
	log.SetLevel(log.Level(uint8(c.Int("log-level"))))
	return nil
}

func printStartMessage(c *cli.Context) error {
	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://docs.loraserver.io/",
	}).Info("starting LoRa App Server")
	return nil
}

func setPostgreSQLConnection(c *cli.Context) error {
	log.Info("connecting to postgresql")
	db, err := storage.OpenDatabase(c.String("postgres-dsn"))
	if err != nil {
		return errors.Wrap(err, "database connection error")
	}
	common.DB = db
	return nil
}

func setRedisPool(c *cli.Context) error {
	// setup redis pool
	log.Info("setup redis connection pool")
	common.RedisPool = storage.NewRedisPool(c.String("redis-url"))
	return nil
}

func setHandler(c *cli.Context) error {
	h, err := handler.NewMQTTHandler(common.RedisPool, c.String("mqtt-server"), c.String("mqtt-username"), c.String("mqtt-password"), c.String("mqtt-ca-cert"))
	if err != nil {
		return errors.Wrap(err, "setup mqtt handler error")
	}
	common.Handler = h
	return nil
}

func setNetworkServerClient(c *cli.Context) error {
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
		return errors.Wrap(err, "network-server dial error")
	}
	common.NetworkServer = ns.NewNetworkServerClient(nsConn)

	return nil
}

func runDatabaseMigrations(c *cli.Context) error {
	if c.Bool("db-automigrate") {
		log.Info("applying database migrations")
		m := &migrate.AssetMigrationSource{
			Asset:    migrations.Asset,
			AssetDir: migrations.AssetDir,
			Dir:      "",
		}
		n, err := migrate.Exec(common.DB.DB, "postgres", m, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "applying migrations error")
		}
		log.WithField("count", n).Info("migrations applied")
	}

	if err := gwmigrate.MigrateGateways(); err != nil {
		log.Fatalf("migrate gateway data error: %s", err)
	}

	return nil
}

func setJWTSecret(c *cli.Context) error {
	storage.SetUserSecret(c.String("jwt-secret"))
	return nil
}

func setHashIterations(c *cli.Context) error {
	storage.HashIterations = c.Int("pw-hash-iterations")
	return nil
}

func setDisableAssignExistingUsers(c *cli.Context) error {
	auth.DisableAssignExistingUsers = c.Bool("disable-assign-existing-users")
	return nil
}

func handleDataDownPayloads(c *cli.Context) error {
	go downlink.HandleDataDownPayloads()
	return nil
}

func startApplicationServerAPI(c *cli.Context) error {
	log.WithFields(log.Fields{
		"bind":     c.String("bind"),
		"ca-cert":  c.String("ca-cert"),
		"tls-cert": c.String("tls-cert"),
		"tls-key":  c.String("tls-key"),
	}).Info("starting application-server api")
	apiServer := mustGetAPIServer(c)
	ln, err := net.Listen("tcp", c.String("bind"))
	if err != nil {
		log.Fatalf("start application-server api listener error: %s", err)
	}
	go apiServer.Serve(ln)
	return nil
}

func startClientAPI(ctx context.Context) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// setup the client API interface
		var validator auth.Validator
		if c.String("jwt-secret") != "" {
			validator = auth.NewJWTValidator(common.DB, "HS256", c.String("jwt-secret"))
		} else {
			log.Fatal("--jwt-secret must be set")
		}

		clientAPIHandler := grpc.NewServer()
		pb.RegisterApplicationServer(clientAPIHandler, api.NewApplicationAPI(validator))
		pb.RegisterDownlinkQueueServer(clientAPIHandler, api.NewDownlinkQueueAPI(validator))
		pb.RegisterNodeServer(clientAPIHandler, api.NewNodeAPI(validator))
		pb.RegisterUserServer(clientAPIHandler, api.NewUserAPI(validator))
		pb.RegisterInternalServer(clientAPIHandler, api.NewInternalUserAPI(validator))
		pb.RegisterGatewayServer(clientAPIHandler, api.NewGatewayAPI(validator))
		pb.RegisterOrganizationServer(clientAPIHandler, api.NewOrganizationAPI(validator))

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

		// start the API server
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

		// setup the HTTP handler
		var err error
		clientHTTPHandler, err = getHTTPHandler(ctx, c)
		if err != nil {
			return err
		}

		return nil
	}
}

func mustGetAPIServer(c *cli.Context) *grpc.Server {
	var opts []grpc.ServerOption
	if c.String("tls-cert") != "" && c.String("tls-key") != "" {
		creds := mustGetTransportCredentials(c.String("tls-cert"), c.String("tls-key"), c.String("ca-cert"), false)
		opts = append(opts, grpc.Creds(creds))
	}
	gs := grpc.NewServer(opts...)
	asAPI := api.NewApplicationServerAPI()
	as.RegisterApplicationServerServer(gs, asAPI)
	return gs
}

func getHTTPHandler(ctx context.Context, c *cli.Context) (http.Handler, error) {
	r := mux.NewRouter()

	// setup json api handler
	jsonHandler, err := getJSONGateway(ctx, c)
	if err != nil {
		return nil, err
	}

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

	return r, nil
}

func getJSONGateway(ctx context.Context, c *cli.Context) (http.Handler, error) {
	// dial options for the grpc-gateway
	b, err := ioutil.ReadFile(c.String("http-tls-cert"))
	if err != nil {
		return nil, errors.Wrap(err, "read http-tls-cert cert error")
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, errors.Wrap(err, "failed to append certificate")
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
		return nil, errors.Wrap(err, "register application handler error")
	}
	if err := pb.RegisterDownlinkQueueHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register downlink queue handler error")
	}
	if err := pb.RegisterNodeHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register node handler error")
	}
	if err := pb.RegisterUserHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register user handler error")
	}
	if err := pb.RegisterInternalHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register internal handler error")
	}
	if err := pb.RegisterGatewayHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register gateway handler error")
	}
	if err := pb.RegisterOrganizationHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register organization handler error")
	}

	return mux, nil
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
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})
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
			Name:   "mqtt-ca-cert",
			Usage:  "mqtt CA certificate file used by the gateway backend (optional)",
			EnvVar: "MQTT_CA_CERT",
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
		cli.BoolFlag{
			Name:   "disable-assign-existing-users",
			Usage:  "when set, existing users can't be re-assigned (to avoid exposure of all users to an organization admin)",
			EnvVar: "DISABLE_ASSIGN_EXISTING_USERS",
		},
	}
	app.Run(os.Args)
}
