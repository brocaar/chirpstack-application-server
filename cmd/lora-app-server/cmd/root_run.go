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
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/brocaar/lora-app-server/api"
	"github.com/brocaar/lora-app-server/internal/api"
	"github.com/brocaar/lora-app-server/internal/api/auth"
	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/brocaar/lora-app-server/internal/downlink"
	"github.com/brocaar/lora-app-server/internal/gwping"
	"github.com/brocaar/lora-app-server/internal/integration"
	"github.com/brocaar/lora-app-server/internal/integration/application"
	"github.com/brocaar/lora-app-server/internal/integration/multi"
	"github.com/brocaar/lora-app-server/internal/migrations"
	"github.com/brocaar/lora-app-server/internal/nsclient"
	"github.com/brocaar/lora-app-server/internal/static"
	"github.com/brocaar/lora-app-server/internal/storage"
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
		setupIntegration,
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
		"docs":    "https://www.loraserver.io/",
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
	config.C.Redis.Pool = storage.NewRedisPool(config.C.Redis.URL, config.C.Redis.MaxIdle, config.C.Redis.IdleTimeout)
	return nil
}

func setupIntegration() error {
	var confs []interface{}

	for _, name := range config.C.ApplicationServer.Integration.Enabled {
		switch name {
		case "aws_sns":
			confs = append(confs, config.C.ApplicationServer.Integration.AWSSNS)
		case "azure_service_bus":
			confs = append(confs, config.C.ApplicationServer.Integration.AzureServiceBus)
		case "mqtt":
			confs = append(confs, config.C.ApplicationServer.Integration.MQTT)
		case "gcp_pub_sub":
			confs = append(confs, config.C.ApplicationServer.Integration.GCPPubSub)
		default:
			return fmt.Errorf("unknown integration type: %s", name)
		}
	}

	mi, err := multi.New(confs)
	if err != nil {
		return errors.Wrap(err, "setup integrations error")
	}
	mi.Add(application.New())
	integration.SetIntegration(mi)

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
		Handler:   api.NewJoinServerAPI(),
		Addr:      config.C.JoinServer.Bind,
		TLSConfig: &tls.Config{},
	}

	if config.C.JoinServer.CACert == "" && config.C.JoinServer.TLSCert == "" && config.C.JoinServer.TLSKey == "" {
		go func() {
			err := server.ListenAndServe()
			log.WithError(err).Error("join-server api error")
		}()
		return nil
	}

	if config.C.JoinServer.CACert != "" {
		caCert, err := ioutil.ReadFile(config.C.JoinServer.CACert)
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
			"ca_cert": config.C.JoinServer.CACert,
		}).Info("join-server is configured with client-certificate authentication")
	}

	go func() {
		err := server.ListenAndServeTLS(config.C.JoinServer.TLSCert, config.C.JoinServer.TLSKey)
		log.WithError(err).Error("join-server api error")
	}()

	return nil
}

func setupCorsHeaders(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", config.C.ApplicationServer.ExternalAPI.CORSAllowOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Grpc-Metadata-Authorization")
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

		rpID, err := uuid.FromString(config.C.ApplicationServer.ID)
		if err != nil {
			return errors.Wrap(err, "application-server id to uuid error")
		}

		clientAPIHandler := grpc.NewServer(gRPCLoggingServerOptions()...)
		pb.RegisterApplicationServiceServer(clientAPIHandler, api.NewApplicationAPI(validator))
		pb.RegisterDeviceQueueServiceServer(clientAPIHandler, api.NewDeviceQueueAPI(validator))
		pb.RegisterDeviceServiceServer(clientAPIHandler, api.NewDeviceAPI(validator))
		pb.RegisterUserServiceServer(clientAPIHandler, api.NewUserAPI(validator))
		pb.RegisterInternalServiceServer(clientAPIHandler, api.NewInternalUserAPI(validator))
		pb.RegisterGatewayServiceServer(clientAPIHandler, api.NewGatewayAPI(validator))
		pb.RegisterGatewayProfileServiceServer(clientAPIHandler, api.NewGatewayProfileAPI(validator))
		pb.RegisterOrganizationServiceServer(clientAPIHandler, api.NewOrganizationAPI(validator))
		pb.RegisterNetworkServerServiceServer(clientAPIHandler, api.NewNetworkServerAPI(validator))
		pb.RegisterServiceProfileServiceServer(clientAPIHandler, api.NewServiceProfileServiceAPI(validator))
		pb.RegisterDeviceProfileServiceServer(clientAPIHandler, api.NewDeviceProfileServiceAPI(validator))
		pb.RegisterMulticastGroupServiceServer(clientAPIHandler, api.NewMulticastGroupAPI(validator, config.C.PostgreSQL.DB, rpID, config.C.NetworkServer.Pool))

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
				if config.C.ApplicationServer.ExternalAPI.CORSAllowOrigin != "" {
					setupCorsHeaders(w, r)
					if r.Method == "OPTIONS" {
						return
					}
				}
				clientHTTPHandler.ServeHTTP(w, r)
			}
		})

		// start the API server
		go func() {
			log.WithFields(log.Fields{
				"bind":     config.C.ApplicationServer.ExternalAPI.Bind,
				"tls-cert": config.C.ApplicationServer.ExternalAPI.TLSCert,
				"tls-key":  config.C.ApplicationServer.ExternalAPI.TLSKey,
			}).Info("starting client api server")

			if config.C.ApplicationServer.ExternalAPI.TLSCert == "" || config.C.ApplicationServer.ExternalAPI.TLSKey == "" {
				log.Fatal(http.ListenAndServe(config.C.ApplicationServer.ExternalAPI.Bind, h2c.NewHandler(handler, &http2.Server{})))
			} else {
				log.Fatal(http.ListenAndServeTLS(
					config.C.ApplicationServer.ExternalAPI.Bind,
					config.C.ApplicationServer.ExternalAPI.TLSCert,
					config.C.ApplicationServer.ExternalAPI.TLSKey,
					h2c.NewHandler(handler, &http2.Server{}),
				))
			}
		}()

		// give the http server some time to start
		time.Sleep(time.Millisecond * 100)

		// setup the HTTP handler
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
	as.RegisterApplicationServerServiceServer(gs, asAPI)
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
	var grpcDialOpts []grpc.DialOption

	if config.C.ApplicationServer.ExternalAPI.TLSCert == "" || config.C.ApplicationServer.ExternalAPI.TLSKey == "" {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	} else {
		b, err := ioutil.ReadFile(config.C.ApplicationServer.ExternalAPI.TLSCert)
		if err != nil {
			return nil, errors.Wrap(err, "read external api tls cert error")
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
	}

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

	if err := pb.RegisterApplicationServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register application handler error")
	}
	if err := pb.RegisterDeviceQueueServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register downlink queue handler error")
	}
	if err := pb.RegisterDeviceServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register node handler error")
	}
	if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register user handler error")
	}
	if err := pb.RegisterInternalServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register internal handler error")
	}
	if err := pb.RegisterGatewayServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register gateway handler error")
	}
	if err := pb.RegisterGatewayProfileServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register gateway-profile handler error")
	}
	if err := pb.RegisterOrganizationServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register organization handler error")
	}
	if err := pb.RegisterNetworkServerServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register network-server handler error")
	}
	if err := pb.RegisterServiceProfileServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register service-profile handler error")
	}
	if err := pb.RegisterDeviceProfileServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register device-profile handler error")
	}
	if err := pb.RegisterMulticastGroupServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register multicast-group handler error")
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
