package cmd

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

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/gusseleet/lora-app-server/api"
	"github.com/gusseleet/lora-app-server/internal/api"
	"github.com/gusseleet/lora-app-server/internal/api/auth"
	"github.com/gusseleet/lora-app-server/internal/config"
	"github.com/gusseleet/lora-app-server/internal/downlink"
	"github.com/gusseleet/lora-app-server/internal/gwping"
	"github.com/gusseleet/lora-app-server/internal/handler/mqtthandler"
	"github.com/gusseleet/lora-app-server/internal/handler/multihandler"
	"github.com/gusseleet/lora-app-server/internal/migrations"
	"github.com/gusseleet/lora-app-server/internal/nsclient"
	"github.com/gusseleet/lora-app-server/internal/profilesmigrate"
	"github.com/gusseleet/lora-app-server/internal/queuemigrate"
	"github.com/gusseleet/lora-app-server/internal/static"
	"github.com/gusseleet/lora-app-server/internal/storage"
	"github.com/brocaar/loraserver/api/as"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
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
		startGatewayPing,
		startJoinServerAPI,
		startClientAPI(ctx),
	}

	for _, t := range tasks {
		if err := t(); err != nil {
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

func setLogLevel() error {
	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))
	return nil
}

func printStartMessage() error {
	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://docs.loraserver.io/",
	}).Info("starting LoRa App Server")
	return nil
}

func setPostgreSQLConnection() error {
	log.Info("connecting to postgresql")
	db, err := storage.OpenDatabase(config.C.PostgreSQL.DSN)
	if err != nil {
		return errors.Wrap(err, "database connection error")
	}
	config.C.PostgreSQL.DB = db
	return nil
}

func setRedisPool() error {
	// setup redis pool
	log.Info("setup redis connection pool")
	config.C.Redis.Pool = storage.NewRedisPool(config.C.Redis.URL)
	return nil
}

func setHandler() error {
	h, err := mqtthandler.NewHandler(
		config.C.ApplicationServer.Integration.MQTT.Server,
		config.C.ApplicationServer.Integration.MQTT.Username,
		config.C.ApplicationServer.Integration.MQTT.Password,
		config.C.ApplicationServer.Integration.MQTT.CACert,
		config.C.ApplicationServer.Integration.MQTT.TLSCert,
		config.C.ApplicationServer.Integration.MQTT.TLSKey,
	)
	if err != nil {
		return errors.Wrap(err, "setup mqtt handler error")
	}
	config.C.ApplicationServer.Integration.Handler = multihandler.NewHandler(h)
	return nil
}

func setNetworkServerClient() error {
	config.C.NetworkServer.Pool = nsclient.NewPool()
	return nil
}

func runDatabaseMigrations() error {
	if config.C.PostgreSQL.Automigrate {
		log.Info("applying database migrations")
		m := &migrate.AssetMigrationSource{
			Asset:    migrations.Asset,
			AssetDir: migrations.AssetDir,
			Dir:      "",
		}
		n, err := migrate.Exec(config.C.PostgreSQL.DB.DB.DB, "postgres", m, migrate.Up)
		if err != nil {
			return errors.Wrap(err, "applying migrations error")
		}
		log.WithField("count", n).Info("migrations applied")

		for {
			if err := profilesmigrate.StartProfilesMigration(config.C.NetworkServer.Server); err != nil {
				log.WithError(err).Error("profiles migration failed")
				time.Sleep(time.Second * 2)
				continue
			}
			break
		}

		for {
			if err := queuemigrate.StartDeviceQueueMigration(); err != nil {
				log.WithError(err).Error("device-queue migration failed")
				time.Sleep(time.Second * 2)
				continue
			}
			break
		}
	}

	return nil
}

func setJWTSecret() error {
	storage.SetUserSecret(config.C.ApplicationServer.ExternalAPI.JWTSecret)
	return nil
}

func setHashIterations() error {
	storage.HashIterations = config.C.General.PasswordHashIterations
	return nil
}

func setDisableAssignExistingUsers() error {
	auth.DisableAssignExistingUsers = config.C.ApplicationServer.ExternalAPI.DisableAssignExistingUsers
	return nil
}

func handleDataDownPayloads() error {
	go downlink.HandleDataDownPayloads()
	return nil
}

func startApplicationServerAPI() error {
	log.WithFields(log.Fields{
		"bind":     config.C.ApplicationServer.API.Bind,
		"ca-cert":  config.C.ApplicationServer.API.CACert,
		"tls-cert": config.C.ApplicationServer.API.TLSCert,
		"tls-key":  config.C.ApplicationServer.API.TLSKey,
	}).Info("starting application-server api")
	apiServer := mustGetAPIServer()
	ln, err := net.Listen("tcp", config.C.ApplicationServer.API.Bind)
	if err != nil {
		log.Fatalf("start application-server api listener error: %s", err)
	}
	go apiServer.Serve(ln)
	return nil
}

func startGatewayPing() error {
	if !config.C.ApplicationServer.GatewayDiscovery.Enabled {
		return nil
	}

	if config.C.ApplicationServer.GatewayDiscovery.Frequency == 0 {
		log.Fatalf("gateway discovery frequency must be set")
	}

	go gwping.SendPingLoop()

	return nil
}

func startJoinServerAPI() error {
	log.WithFields(log.Fields{
		"bind":     config.C.JoinServer.Bind,
		"ca_cert":  config.C.JoinServer.CACert,
		"tls_cert": config.C.JoinServer.TLSCert,
		"tls_key":  config.C.JoinServer.TLSKey,
	}).Info("starting join-server api")

	server := http.Server{
		Handler: api.NewJoinServerAPI(),
		Addr:    config.C.JoinServer.Bind,
	}

	if config.C.JoinServer.CACert == "" || config.C.JoinServer.TLSCert == "" || config.C.JoinServer.TLSKey == "" {
		go func() {
			err := server.ListenAndServe()
			log.WithError(err).Error("join-server api error")
		}()
		return nil
	}

	caCert, err := ioutil.ReadFile(config.C.JoinServer.CACert)
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

	go func() {
		err := server.ListenAndServeTLS(config.C.JoinServer.TLSCert, config.C.JoinServer.TLSKey)
		log.WithError(err).Error("join-server api error")
	}()

	return nil
}

func startClientAPI(ctx context.Context) func() error {
	return func() error {
		// setup the client API interface
		var validator auth.Validator
		if config.C.ApplicationServer.ExternalAPI.JWTSecret != "" {
			validator = auth.NewJWTValidator(config.C.PostgreSQL.DB, "HS256", config.C.ApplicationServer.ExternalAPI.JWTSecret)
		} else {
			log.Fatal("jwt secret must be set for external api")
		}

		clientAPIHandler := grpc.NewServer(gRPCLoggingServerOptions()...)
		pb.RegisterApplicationServer(clientAPIHandler, api.NewApplicationAPI(validator))
		pb.RegisterDeviceQueueServer(clientAPIHandler, api.NewDeviceQueueAPI(validator))
		pb.RegisterDeviceServer(clientAPIHandler, api.NewDeviceAPI(validator))
		pb.RegisterUserServer(clientAPIHandler, api.NewUserAPI(validator))
		pb.RegisterInternalServer(clientAPIHandler, api.NewInternalUserAPI(validator))
		pb.RegisterGatewayServer(clientAPIHandler, api.NewGatewayAPI(validator))
		pb.RegisterOrganizationServer(clientAPIHandler, api.NewOrganizationAPI(validator))
		pb.RegisterNetworkServerServer(clientAPIHandler, api.NewNetworkServerAPI(validator))
		pb.RegisterServiceProfileServiceServer(clientAPIHandler, api.NewServiceProfileServiceAPI(validator))
		pb.RegisterDeviceProfileServiceServer(clientAPIHandler, api.NewDeviceProfileServiceAPI(validator))

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
			if config.C.ApplicationServer.ExternalAPI.TLSCert == "" || config.C.ApplicationServer.ExternalAPI.TLSKey == "" {
				log.Fatal("tls cert and tls key must be set for the external api")
			}
			log.WithFields(log.Fields{
				"bind":     config.C.ApplicationServer.ExternalAPI.Bind,
				"tls-cert": config.C.ApplicationServer.ExternalAPI.TLSCert,
				"tls-key":  config.C.ApplicationServer.ExternalAPI.TLSKey,
			}).Info("starting client api server")
			log.Fatal(http.ListenAndServeTLS(config.C.ApplicationServer.ExternalAPI.Bind, config.C.ApplicationServer.ExternalAPI.TLSCert, config.C.ApplicationServer.ExternalAPI.TLSKey, handler))
		}()

		// give the http server some time to start
		time.Sleep(time.Millisecond * 100)

		// setup the HTTP handler
		var err error
		clientHTTPHandler, err = getHTTPHandler(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

func gRPCLoggingServerOptions() []grpc.ServerOption {
	logrusEntry := log.NewEntry(log.StandardLogger())
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry, logrusOpts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry, logrusOpts...),
		),
	}
}

func mustGetAPIServer() *grpc.Server {
	opts := gRPCLoggingServerOptions()
	if config.C.ApplicationServer.API.CACert != "" && config.C.ApplicationServer.API.TLSCert != "" && config.C.ApplicationServer.API.TLSKey != "" {
		creds := mustGetTransportCredentials(config.C.ApplicationServer.API.TLSCert, config.C.ApplicationServer.API.TLSKey, config.C.ApplicationServer.API.CACert, true)
		opts = append(opts, grpc.Creds(creds))
	}
	gs := grpc.NewServer(opts...)
	asAPI := api.NewApplicationServerAPI()
	as.RegisterApplicationServerServer(gs, asAPI)
	return gs
}

func getHTTPHandler(ctx context.Context) (http.Handler, error) {
	r := mux.NewRouter()

	// setup json api handler
	jsonHandler, err := getJSONGateway(ctx)
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

	return wsproxy.WebsocketProxy(r), nil
}

func getJSONGateway(ctx context.Context) (http.Handler, error) {
	// dial options for the grpc-gateway
	b, err := ioutil.ReadFile(config.C.ApplicationServer.ExternalAPI.TLSCert)
	if err != nil {
		return nil, errors.Wrap(err, "read external api tls cert error")
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

	bindParts := strings.SplitN(config.C.ApplicationServer.ExternalAPI.Bind, ":", 2)
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
	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		log.WithFields(log.Fields{
			"cert": tlsCert,
			"key":  tlsKey,
		}).Fatalf("load key-pair error: %s", err)
	}

	rawCaCert, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.WithField("ca", caCert).Fatalf("load ca cert error: %s", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(rawCaCert)

	if verifyClientCert {
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientCAs:    caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})
}
