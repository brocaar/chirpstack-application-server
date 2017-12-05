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

	"github.com/brocaar/lora-app-server/internal/nsclient"
	"github.com/brocaar/lora-app-server/internal/profilesmigrate"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/common"
	"github.com/brocaar/lora-app-server/internal/downlink"
	"github.com/brocaar/lora-app-server/internal/gwping"
	"github.com/brocaar/lora-app-server/internal/handler/mqtthandler"
	"github.com/brocaar/lora-app-server/internal/handler/multihandler"
	"github.com/brocaar/lora-app-server/internal/migrations"
	"github.com/brocaar/lora-app-server/internal/static"
	"github.com/brocaar/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/as"
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
		setPublicASSettings,
		handleDataDownPayloads,
		startApplicationServerAPI,
		startGatewayPing,
		startJoinServerAPI,
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
	h, err := mqtthandler.NewHandler(c.String("mqtt-server"), c.String("mqtt-username"), c.String("mqtt-password"), c.String("mqtt-ca-cert"))
	if err != nil {
		return errors.Wrap(err, "setup mqtt handler error")
	}
	common.Handler = multihandler.NewHandler(h)
	return nil
}

func setNetworkServerClient(c *cli.Context) error {
	common.NetworkServerPool = nsclient.NewPool(
		c.String("ns-ca-cert"),
		c.String("ns-tls-cert"),
		c.String("ns-tls-key"),
	)

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

		for {
			if err := profilesmigrate.StartProfilesMigration(c.String("ns-server")); err != nil {
				log.WithError(err).Error("profiles migration failed")
				time.Sleep(time.Second * 2)
				continue
			}
			break
		}
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

func setPublicASSettings(c *cli.Context) error {
	// TODO: get from client-side certificate in the future?
	common.ApplicationServerID = c.String("as-public-id")
	common.ApplicationServerServer = c.String("as-public-server")
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

func startGatewayPing(c *cli.Context) error {
	if !c.Bool("gw-ping") {
		return nil
	}

	common.GatewayPingFrequency = c.Int("gw-ping-frequency")
	common.GatewayPingDR = c.Int("gw-ping-dr")
	common.GatewayPingInterval = c.Duration("gw-ping-interval")

	if common.GatewayPingFrequency == 0 || common.GatewayPingDR == 0 {
		log.Fatalf("--gw-ping-frequency and --gw-ping-dr settings must be set")
	}

	go gwping.SendPingLoop()

	return nil
}

func startJoinServerAPI(c *cli.Context) error {
	log.WithFields(log.Fields{
		"bind":     c.String("js-bind"),
		"ca_cert":  c.String("js-ca-cert"),
		"tls_cert": c.String("js-tls-cert"),
		"tls_key":  c.String("js-tls-key"),
	}).Info("starting join-server api")

	server := http.Server{
		Handler: api.NewJoinServerAPI(),
		Addr:    c.String("js-bind"),
	}

	if c.String("js-ca-cert") != "" {
		caCert, err := ioutil.ReadFile(c.String("js-ca-cert"))
		if err != nil {
			return errors.Wrap(err, "read ca certificate error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return errors.New("append ca certificate error")
		}

		server.TLSConfig = &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
	}

	if c.String("js-tls-cert") == "" && c.String("js-tls-key") == "" {
		go func() {
			err := server.ListenAndServe()
			log.WithError(err).Error("join-server api error")
		}()
	} else {
		go func() {
			err := server.ListenAndServeTLS(c.String("js-tls-cert"), c.String("js-tls-key"))
			log.WithError(err).Error("join-server api error")
		}()
	}

	return nil
}

func startClientAPI(ctx context.Context) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// Get client gRPC API
		clientAPIHandler := getGRPCClientAPI(c)

		// setup the client http interface variable
		// we need to start the gRPC service first, as it is used by the
		// grpc-gateway
		var clientHTTPHandler http.Handler

		// switch between gRPC and "plain" http handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c.String("grpc-bind") == "" && r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				clientAPIHandler.ServeHTTP(w, r)
			} else {
				if clientHTTPHandler == nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				clientHTTPHandler.ServeHTTP(w, r)
			}
		})

		// If gRPC API is to share port with the normal http service,
		// ie use ServeHTTP, it is currently necessary to use TLS.
		// See https://github.com/grpc/grpc-go/issues/555 for more info.
		if c.String("grpc-bind") == "" && (c.String("http-tls-cert") == "" || c.String("http-tls-key") == "") {
			log.Fatal("--http-tls-cert (HTTP_TLS_CERT) and --http-tls-key (HTTP_TLS_KEY) must be set")
		}

		// If grpc-bind is set we should serve gRPC separate from the normal http handler
		if c.String("grpc-bind") != "" {
			ln, err := net.Listen("tcp", c.String("grpc-bind"))
			if err != nil {
				log.Fatalf("start grpc-server api listener error: %s", err)
			}

			log.WithFields(log.Fields{
				"bind":     c.String("grpc-bind"),
				"tls-cert": c.String("grpc-tls-cert"),
				"tls-key":  c.String("grpc-tls-key"),
			}).Info("starting client grpc api server")

			go clientAPIHandler.Serve(ln)
		}

		// start the API server
		go func() {
			log.WithFields(log.Fields{
				"bind":     c.String("http-bind"),
				"tls-cert": c.String("http-tls-cert"),
				"tls-key":  c.String("http-tls-key"),
			}).Info("starting client api server")

			if c.String("http-tls-cert") == "" && c.String("http-tls-key") == "" {
				log.Fatal(http.ListenAndServe(c.String("http-bind"), handler))
			} else {
				log.Fatal(http.ListenAndServeTLS(c.String("http-bind"), c.String("http-tls-cert"),
					c.String("http-tls-key"), handler))
			}
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

func getGRPCClientAPI(c *cli.Context) *grpc.Server {
	var validator auth.Validator
	if c.String("jwt-secret") != "" {
		validator = auth.NewJWTValidator(common.DB, "HS256", c.String("jwt-secret"))
	} else {
		log.Fatal("--jwt-secret must be set")
	}

	// Only use grpc-tls-cert and grpc-tls-key if we do not share port with http service
	var opts []grpc.ServerOption
	if c.String("grpc-bind") != "" && c.String("grpc-tls-cert") != "" && c.String("grpc-tls-key") != "" {
		creds := mustGetTransportCredentials(c.String("grpc-tls-cert"), c.String("grpc-tls-key"), "", false)
		opts = append(opts, grpc.Creds(creds))
	}

	gs := grpc.NewServer(opts...)
	pb.RegisterApplicationServer(gs, api.NewApplicationAPI(validator))
	pb.RegisterDeviceQueueServer(gs, api.NewDeviceQueueAPI(validator))
	pb.RegisterDeviceServer(gs, api.NewDeviceAPI(validator))
	pb.RegisterUserServer(gs, api.NewUserAPI(validator))
	pb.RegisterInternalServer(gs, api.NewInternalUserAPI(validator, c))
	pb.RegisterGatewayServer(gs, api.NewGatewayAPI(validator))
	pb.RegisterOrganizationServer(gs, api.NewOrganizationAPI(validator))
	pb.RegisterNetworkServerServer(gs, api.NewNetworkServerAPI(validator))
	pb.RegisterServiceProfileServiceServer(gs, api.NewServiceProfileServiceAPI(validator))
	pb.RegisterDeviceProfileServiceServer(gs, api.NewDeviceProfileServiceAPI(validator))

	return gs
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
	var bind, cert string

	if c.String("grpc-bind") != "" {
		bind = "grpc-bind"
		if c.String("grpc-tls-cert") != "" {
			cert = "grpc-tls-cert"
		}
	} else {
		bind = "http-bind"
		cert = "http-tls-cert"
	}

	var grpcDialOpts []grpc.DialOption

	// setup grpc-gateway dial options
	if cert != "" {
		b, err := ioutil.ReadFile(c.String(cert))
		if err != nil {
			return nil, errors.Wrapf(err, "read %s cert error", cert)
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			return nil, errors.Wrap(err, "failed to append certificate")
		}
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			// given the grpc-gateway is always connecting to localhost, does
			// InsecureSkipVerify=true cause any security issues?
			InsecureSkipVerify: true,
			RootCAs:            cp,
		})))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}

	bindParts := strings.SplitN(c.String(bind), ":", 2)
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
	if err := pb.RegisterDeviceQueueHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register downlink queue handler error")
	}
	if err := pb.RegisterDeviceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
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
	if err := pb.RegisterNetworkServerHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register network-server handler error")
	}
	if err := pb.RegisterServiceProfileServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register service-profile handler error")
	}
	if err := pb.RegisterDeviceProfileServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register device-profile handler error")
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
			Name:   "as-public-server",
			Usage:  "ip:port of the application-server api (used by LoRa Server to connect back to LoRa App Server)",
			Value:  "localhost:8001",
			EnvVar: "AS_PUBLIC_SERVER",
		},
		cli.StringFlag{
			Name:   "as-public-id",
			Usage:  "random uuid defining the id of the application-server installation (used by LoRa Server as routing-profile id)",
			Value:  "6d5db27e-4ce2-4b2b-b5d7-91f069397978",
			EnvVar: "AS_PUBLIC_ID",
		},
		cli.StringFlag{
			Name:   "bind",
			Usage:  "ip:port to bind the api server",
			Value:  "0.0.0.0:8001",
			EnvVar: "BIND",
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
		cli.BoolFlag{
			Name:   "gw-ping",
			Usage:  "enable sending gateway pings",
			EnvVar: "GW_PING",
		},
		cli.DurationFlag{
			Name:   "gw-ping-interval",
			Usage:  "the interval used for each gateway to send a ping",
			EnvVar: "GW_PING_INTERVAL",
			Value:  time.Hour * 24,
		},
		cli.IntFlag{
			Name:   "gw-ping-frequency",
			Usage:  "the frequency used for transmitting the gateway ping (in Hz)",
			EnvVar: "GW_PING_FREQUENCY",
		},
		cli.IntFlag{
			Name:   "gw-ping-dr",
			Usage:  "the data-rate to use for transmitting the gateway ping",
			EnvVar: "GW_PING_DR",
		},
		cli.StringFlag{
			Name:   "branding-header",
			Usage:  "when set, this html is inserted into the header of the ui, before \"LoRa Server\"",
			EnvVar: "BRANDING_HEADER",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "branding-footer",
			Usage:  "when set, this html is inserted as a footer of the ui pages",
			EnvVar: "BRANDING_FOOTER",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "branding-registration",
			Usage:  "when set, this html is inserted onto the login page, under the login area",
			EnvVar: "BRANDING_REGISTRATION",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "js-bind",
			Usage:  "ip:port to bind the join-server api interface to",
			Value:  "0.0.0.0:8003",
			EnvVar: "JS_BIND",
		},
		cli.StringFlag{
			Name:   "js-ca-cert",
			Usage:  "ca certificate used by the join-server api server (optional)",
			EnvVar: "JS_CA_CERT",
		},
		cli.StringFlag{
			Name:   "js-tls-cert",
			Usage:  "tls certificate used by the join-server api server (optional)",
			EnvVar: "JS_TLS_CERT",
		},
		cli.StringFlag{
			Name:   "js-tls-key",
			Usage:  "tls key used by the join-server api server (optional)",
			EnvVar: "JS_TLS_KEY",
		},
		cli.StringFlag{
			Name:   "ns-server",
			Usage:  "hostname:port of the network-server api server",
			Value:  "127.0.0.1:8000",
			EnvVar: "NS_SERVER",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "grpc-bind",
			Usage:  "hostname:port to bind gRPC client API to, will use http-bind if unset (optional)",
			EnvVar: "GRPC_BIND",
		},
		cli.StringFlag{
			Name:   "grpc-tls-cert",
			Usage:  "tls certificate used by gRPC client API, only used if grpc-bind is set (optional)",
			EnvVar: "GRPC_TLS_CERT",
		},
		cli.StringFlag{
			Name:   "grpc-tls-key",
			Usage:  "tls key used by gRPC client API, only used if grpc-bind is set (optional)",
			EnvVar: "GRPC_TLS_KEY",
		},
	}
	app.Run(os.Args)
}
