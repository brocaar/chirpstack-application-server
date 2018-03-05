package cmd

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gusseleet/lora-app-server/internal/config"
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
	> documentation & support: https://docs.loraserver.io/lora-app-server
	> source & copyright information: https://github.com/brocaar/lora-app-server`,
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 4, "debug=5, info=4, error=2, fatal=1, panic=0")

	// for backwards compatibility
	rootCmd.PersistentFlags().String("postgres-dsn", "postgres://localhost/loraserver_as?sslmode=disable", "")
	rootCmd.PersistentFlags().Bool("db-automigrate", true, "")
	rootCmd.PersistentFlags().String("redis-url", "redis://localhost:6379", "")
	rootCmd.PersistentFlags().String("mqtt-server", "tcp://localhost:1883", "")
	rootCmd.PersistentFlags().String("mqtt-username", "", "")
	rootCmd.PersistentFlags().String("mqtt-password", "", "")
	rootCmd.PersistentFlags().String("mqtt-ca-cert", "", "")
	rootCmd.PersistentFlags().String("mqtt-tls-cert", "", "")
	rootCmd.PersistentFlags().String("mqtt-tls-key", "", "")
	rootCmd.PersistentFlags().String("as-public-server", "localhost:8001", "")
	rootCmd.PersistentFlags().String("as-public-id", "6d5db27e-4ce2-4b2b-b5d7-91f069397978", "")
	rootCmd.PersistentFlags().String("bind", "0.0.0.0:8001", "")
	rootCmd.PersistentFlags().String("ca-cert", "", "")
	rootCmd.PersistentFlags().String("tls-cert", "", "")
	rootCmd.PersistentFlags().String("tls-key", "", "")
	rootCmd.PersistentFlags().String("http-bind", "0.0.0.0:8080", "")
	rootCmd.PersistentFlags().String("http-tls-cert", "", "")
	rootCmd.PersistentFlags().String("http-tls-key", "", "")
	rootCmd.PersistentFlags().String("jwt-secret", "", "")
	rootCmd.PersistentFlags().Int("pw-hash-iterations", 100000, "")
	rootCmd.PersistentFlags().Bool("disable-assign-existing-users", false, "")
	rootCmd.PersistentFlags().Bool("gw-ping", false, "")
	rootCmd.PersistentFlags().Duration("gw-ping-interval", time.Hour*24, "")
	rootCmd.PersistentFlags().Int("gw-ping-frequency", 868100000, "")
	rootCmd.PersistentFlags().Int("gw-ping-dr", 5, "")
	rootCmd.PersistentFlags().String("branding-header", "", "")
	rootCmd.PersistentFlags().String("branding-footer", "", "")
	rootCmd.PersistentFlags().String("branding-registration", "", "")
	rootCmd.PersistentFlags().String("js-bind", "0.0.0.0:8003", "")
	rootCmd.PersistentFlags().String("js-ca-cert", "", "")
	rootCmd.PersistentFlags().String("js-tls-cert", "", "")
	rootCmd.PersistentFlags().String("js-tls-key", "", "")
	rootCmd.PersistentFlags().String("ns-server", "127.0.0.1:8000", "")

	hidden := []string{
		"postgres-dsn", "db-automigrate", "redis-url", "mqtt-server", "mqtt-username", "mqtt-password",
		"mqtt-ca-cert", "mqtt-tls-cert", "mqtt-tls-key", "as-public-server", "as-public-id", "bind", "ca-cert", "tls-cert", "tls-key",
		"http-bind", "http-tls-cert", "http-tls-key", "jwt-secret", "pw-hash-iterations", "disable-assign-existing-users",
		"gw-ping", "gw-ping-interval", "gw-ping-frequency", "gw-ping-dr", "branding-header", "branding-footer", "branding-registration",
		"js-bind", "js-ca-cert", "js-tls-cert", "js-tls-key", "ns-server",
	}
	for _, key := range hidden {
		rootCmd.PersistentFlags().MarkHidden(key)
	}

	viper.BindEnv("general.log_level", "LOG_LEVEL")

	// for backwards compatibility
	viper.BindEnv("general.password_hash_iterations", "PW_HASH_ITERATIONS")
	viper.BindEnv("postgresql.dsn", "POSTGRES_DSN")
	viper.BindEnv("postgresql.automigrate", "DB_AUTOMIGRATE")
	viper.BindEnv("redis.url", "REDIS_URL")
	viper.BindEnv("application_server.id", "AS_PUBLIC_ID")
	viper.BindEnv("application_server.integration.mqtt.server", "MQTT_SERVER")
	viper.BindEnv("application_server.integration.mqtt.username", "MQTT_USERNAME")
	viper.BindEnv("application_server.integration.mqtt.password", "MQTT_PASSWORD")
	viper.BindEnv("application_server.integration.mqtt.ca_cert", "MQTT_CA_CERT")
	viper.BindEnv("application_server.backend.mqtt.tls_cert", "MQTT_TLS_CERT")
	viper.BindEnv("application_server.backend.mqtt.tls_key", "MQTT_TLS_KEY")
	viper.BindEnv("application_server.api.bind", "BIND")
	viper.BindEnv("application_server.api.ca_cert", "CA_CERT")
	viper.BindEnv("application_server.api.tls_cert", "TLS_CERT")
	viper.BindEnv("application_server.api.tls_key", "TLS_KEY")
	viper.BindEnv("application_server.api.public_host", "AS_PUBLIC_SERVER")
	viper.BindEnv("application_server.external_api.bind", "HTTP_BIND")
	viper.BindEnv("application_server.external_api.tls_cert", "HTTP_TLS_CERT")
	viper.BindEnv("application_server.external_api.tls_key", "HTTP_TLS_KEY")
	viper.BindEnv("application_server.external_api.jwt_secret", "JWT_SECRET")
	viper.BindEnv("application_server.external_api.disable_assign_existing_users", "DISABLE_ASSIGN_EXISTING_USERS")
	viper.BindEnv("application_server.branding.header", "BRANDING_HEADER")
	viper.BindEnv("application_server.branding.footer", "BRANDING_FOOTER")
	viper.BindEnv("application_server.branding.registration", "BRANDING_REGISTRATION")
	viper.BindEnv("application_server.gateway_discovery.enabled", "GW_PING")
	viper.BindEnv("application_server.gateway_discovery.interval", "GW_PING_INTERVAL")
	viper.BindEnv("application_server.gateway_discovery.frequency", "GW_PING_FREQUENCY")
	viper.BindEnv("application_server.gateway_discovery.dr", "GW_PING_DR")
	viper.BindEnv("join_server.bind", "JS_BIND")
	viper.BindEnv("join_server.ca_cert", "JS_CA_CERT")
	viper.BindEnv("join_server.tls_cert", "JS_TLS_CERT")
	viper.BindEnv("join_server.tls_key", "JS_TLS_KEY")
	viper.BindEnv("network_server.server", "NS_SERVER")

	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// for backwards compatibility
	viper.BindPFlag("general.password_hash_iterations", rootCmd.PersistentFlags().Lookup("pw-hash-iterations"))
	viper.BindPFlag("postgresql.dsn", rootCmd.PersistentFlags().Lookup("postgres-dsn"))
	viper.BindPFlag("postgresql.automigrate", rootCmd.PersistentFlags().Lookup("db-automigrate"))
	viper.BindPFlag("redis.url", rootCmd.PersistentFlags().Lookup("redis-url"))
	viper.BindPFlag("application_server.id", rootCmd.PersistentFlags().Lookup("as-public-id"))
	viper.BindPFlag("application_server.integration.mqtt.server", rootCmd.PersistentFlags().Lookup("mqtt-server"))
	viper.BindPFlag("application_server.integration.mqtt.username", rootCmd.PersistentFlags().Lookup("mqtt-username"))
	viper.BindPFlag("application_server.integration.mqtt.password", rootCmd.PersistentFlags().Lookup("mqtt-password"))
	viper.BindPFlag("application_server.integration.mqtt.ca_cert", rootCmd.PersistentFlags().Lookup("mqtt-ca-cert"))
	viper.BindPFlag("application_server.backend.mqtt.tls_cert", rootCmd.PersistentFlags().Lookup("mqtt-tls-cert"))
	viper.BindPFlag("application_server.backend.mqtt.tls_key", rootCmd.PersistentFlags().Lookup("mqtt-tls-key"))
	viper.BindPFlag("application_server.api.bind", rootCmd.PersistentFlags().Lookup("bind"))
	viper.BindPFlag("application_server.api.ca_cert", rootCmd.PersistentFlags().Lookup("ca-cert"))
	viper.BindPFlag("application_server.api.tls_cert", rootCmd.PersistentFlags().Lookup("tls-cert"))
	viper.BindPFlag("application_server.api.tls_key", rootCmd.PersistentFlags().Lookup("tls-key"))
	viper.BindPFlag("application_server.api.public_host", rootCmd.PersistentFlags().Lookup("as-public-server"))
	viper.BindPFlag("application_server.external_api.bind", rootCmd.PersistentFlags().Lookup("http-bind"))
	viper.BindPFlag("application_server.external_api.tls_cert", rootCmd.PersistentFlags().Lookup("http-tls-cert"))
	viper.BindPFlag("application_server.external_api.tls_key", rootCmd.PersistentFlags().Lookup("http-tls-key"))
	viper.BindPFlag("application_server.external_api.jwt_secret", rootCmd.PersistentFlags().Lookup("jwt-secret"))
	viper.BindPFlag("application_server.external_api.disable_assign_existing_users", rootCmd.PersistentFlags().Lookup("disable-assign-existing-users"))
	viper.BindPFlag("application_server.branding.header", rootCmd.PersistentFlags().Lookup("branding-header"))
	viper.BindPFlag("application_server.branding.footer", rootCmd.PersistentFlags().Lookup("branding-footer"))
	viper.BindPFlag("application_server.branding.registration", rootCmd.PersistentFlags().Lookup("branding-registration"))
	viper.BindPFlag("application_server.gateway_discovery.enabled", rootCmd.PersistentFlags().Lookup("gw-ping"))
	viper.BindPFlag("application_server.gateway_discovery.interval", rootCmd.PersistentFlags().Lookup("gw-ping-interval"))
	viper.BindPFlag("application_server.gateway_discovery.frequency", rootCmd.PersistentFlags().Lookup("gw-ping-frequency"))
	viper.BindPFlag("application_server.gateway_discovery.dr", rootCmd.PersistentFlags().Lookup("gw-ping-dr"))
	viper.BindPFlag("join_server.bind", rootCmd.PersistentFlags().Lookup("js-bind"))
	viper.BindPFlag("join_server.ca_cert", rootCmd.PersistentFlags().Lookup("js-ca-cert"))
	viper.BindPFlag("join_server.tls_cert", rootCmd.PersistentFlags().Lookup("js-tls-cert"))
	viper.BindPFlag("join_server.tls_key", rootCmd.PersistentFlags().Lookup("js-tls-key"))
	viper.BindPFlag("network_server.server", rootCmd.PersistentFlags().Lookup("ns-server"))

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
				log.Warning("Deprecation warning! no configuration file found, falling back on environment variables. Update your configuration, see: https://docs.loraserver.io/lora-app-server/install/config/")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}
