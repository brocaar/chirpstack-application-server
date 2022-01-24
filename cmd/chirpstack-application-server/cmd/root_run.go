package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/resolver"

	"github.com/brocaar/chirpstack-application-server/internal/api"
	"github.com/brocaar/chirpstack-application-server/internal/backend/networkserver"
	jscodec "github.com/brocaar/chirpstack-application-server/internal/codec/js"
	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/brocaar/chirpstack-application-server/internal/downlink"
	"github.com/brocaar/chirpstack-application-server/internal/gwping"
	"github.com/brocaar/chirpstack-application-server/internal/integration"
	"github.com/brocaar/chirpstack-application-server/internal/migrations/code"
	"github.com/brocaar/chirpstack-application-server/internal/monitoring"
	"github.com/brocaar/chirpstack-application-server/internal/storage"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tasks := []func() error{
		setLogLevel,
		setSyslog,
		setGRPCResolver,
		printStartMessage,
		setupStorage,
		setupNetworkServer,
		migrateGatewayStats,
		migrateToClusterKeys,
		setupIntegration,
		setupCodec,
		handleDataDownPayloads,
		startGatewayPing,
		setupAPI,
		setupMonitoring,
	}

	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}

	sigChan := make(chan os.Signal, 1)
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

func setGRPCResolver() error {
	resolver.SetDefaultScheme(config.C.General.GRPCDefaultResolverScheme)
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
	if err := integration.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup integration error")
	}

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

func handleDataDownPayloads() error {
	downChan := integration.ForApplicationID(0).DataDownChan()
	go downlink.HandleDataDownPayloads(downChan)
	return nil
}

func migrateGatewayStats() error {
	if err := storage.CodeMigration("migrate_gw_stats", code.MigrateGatewayStats); err != nil {
		return errors.Wrap(err, "migration error")
	}

	return nil
}

func migrateToClusterKeys() error {
	return storage.CodeMigration("migrate_to_cluster_keys", func(db sqlx.Ext) error {
		return code.MigrateToClusterKeys(config.C)
	})
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

func setupMonitoring() error {
	if err := monitoring.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup monitoring error")
	}
	return nil
}
