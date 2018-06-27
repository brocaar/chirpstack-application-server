package cmd

import (
	"bytes"
	"io/ioutil"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string
var version string

var rootCmd = &cobra.Command{
	Use:   "lora-app-server",
	Short: "LoRa Server project application-server",
	Long: `LoRa App Server is an open-source application-server, part of the LoRa Server project
	> documentation & support: https://www.loraserver.io/lora-app-server
	> source & copyright information: https://github.com/brocaar/lora-app-server`,
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 4, "debug=5, info=4, error=2, fatal=1, panic=0")

	// bind flag to config vars
	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// for debian install script
	viper.BindEnv("application_server.external_api.tls_cert", "HTTP_TLS_CERT")
	viper.BindEnv("application_server.external_api.tls_key", "HTTP_TLS_KEY")

	// defaults
	viper.SetDefault("general.password_hash_iterations", 100000)
	viper.SetDefault("postgresql.dsn", "postgres://localhost/loraserver_as?sslmode=disable")
	viper.SetDefault("postgresql.automigrate", true)
	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("application_server.integration.mqtt.server", "tcp://localhost:1883")
	viper.SetDefault("application_server.api.public_host", "localhost:8001")
	viper.SetDefault("application_server.id", "6d5db27e-4ce2-4b2b-b5d7-91f069397978")
	viper.SetDefault("application_server.api.bind", "0.0.0.0:8001")
	viper.SetDefault("application_server.external_api.bind", "0.0.0.0:8080")
	viper.SetDefault("join_server.bind", "0.0.0.0:8003")
	viper.SetDefault("application_server.integration.mqtt.uplink_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/rx")
	viper.SetDefault("application_server.integration.mqtt.downlink_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/tx")
	viper.SetDefault("application_server.integration.mqtt.join_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/join")
	viper.SetDefault("application_server.integration.mqtt.ack_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/ack")
	viper.SetDefault("application_server.integration.mqtt.error_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/error")
	viper.SetDefault("application_server.integration.mqtt.status_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/status")
	viper.SetDefault("application_server.integration.mqtt.clean_session", true)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

// Execute executes the root command.
func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
		viper.SetConfigType("toml")
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("lora-app-server")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/lora-app-server")
		viper.AddConfigPath("/etc/lora-app-server")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No configuration file found, using defaults. See: https://www.loraserver.io/lora-app-server/install/config/")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}
