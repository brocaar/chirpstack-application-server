package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/brocaar/chirpstack-application-server/internal/api"
	"github.com/brocaar/chirpstack-application-server/internal/applayer/fragmentation"
	"github.com/brocaar/chirpstack-application-server/internal/applayer/multicastsetup"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	jscodec "github.com/brocaar/chirpstack-application-server/internal/codec/js"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/downlink"
	"github.com/brocaar/chirpstack-application-server/internal/fuota"
	"github.com/brocaar/chirpstack-application-server/internal/gwping"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/integration/application"
	"github.com/brocaar/chirpstack-application-server/internal/integration/marshaler"
	"github.com/brocaar/chirpstack-application-server/internal/integration/multi"
	"github.com/brocaar/chirpstack-application-server/internal/metrics"
	"github.com/brocaar/chirpstack-application-server/internal/migrations/code"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
		setLogLevel,
		printStartMessage,
		setupStorage,
		setupNetworkServer,
		migrateGatewayStats,
		setupIntegration,
		setupCodec,
		handleDataDownPayloads,
		startGatewayPing,
		setupMulticastSetup,
		setupFragmentation,
		setupFUOTA,
		setupAPI,
		setupMetrics,
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
		log.Warning("stopping chirpstack-application-server")
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
		"docs":    "https://www.chirpstack.io/",
	}).Info("starting ChirpStack Application Server")
	return nil
}

func setupStorage() error {
	if err := storage.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup storage error")
	}

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
		case "postgresql":
			confs = append(confs, config.C.ApplicationServer.Integration.PostgreSQL)
		case "amqp":
			confs = append(confs, config.C.ApplicationServer.Integration.AMQP)
		default:
			return fmt.Errorf("unknown integration type: %s", name)
		}
	}

	var m marshaler.Type
	switch config.C.ApplicationServer.Integration.Marshaler {
	case "protobuf":
		m = marshaler.Protobuf
	case "json":
		m = marshaler.ProtobufJSON
	case "json_v3":
		m = marshaler.JSONV3
	}

	mi, err := multi.New(m, confs)
	if err != nil {
		return errors.Wrap(err, "setup integrations error")
	}
	mi.Add(application.New(m))
	integration.SetIntegration(mi)

	return nil
}

func setupCodec() error {
	if err := jscodec.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup codec error")
	}
	return nil
}

func setupNetworkServer() error {
	if err := networkserver.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup networkserver error")
	}
	return nil
}

func migrateGatewayStats() error {
	if err := code.Migrate("migrate_gw_stats", code.MigrateGatewayStats); err != nil {
		return errors.Wrap(err, "migration error")
	}

	return nil
}

func handleDataDownPayloads() error {
	go downlink.HandleDataDownPayloads()
	return nil
}

func setupAPI() error {
	if err := api.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup api error")
	}
	return nil
}

func startGatewayPing() error {
	go gwping.SendPingLoop()

	return nil
}

func setupMulticastSetup() error {
	if err := multicastsetup.Setup(config.C); err != nil {
		return errors.Wrap(err, "multicastsetup setup error")
	}
	return nil
}

func setupFragmentation() error {
	if err := fragmentation.Setup(config.C); err != nil {
		return errors.Wrap(err, "fragmentation setup error")
	}
	return nil
}

func setupFUOTA() error {
	if err := fuota.Setup(config.C); err != nil {
		return errors.Wrap(err, "fuota setup error")
	}
	return nil
}

func setupMetrics() error {
	if err := metrics.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup metrics error")
	}
	return nil
}
