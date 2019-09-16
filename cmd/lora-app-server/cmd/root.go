package cmd

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

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

	// defaults
	viper.SetDefault("general.password_hash_iterations", 100000)
	viper.SetDefault("postgresql.dsn", "postgres://localhost/loraserver_as?sslmode=disable")
	viper.SetDefault("postgresql.automigrate", true)
	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.max_idle", 10)
	viper.SetDefault("redis.idle_timeout", 5*time.Minute)
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
	viper.SetDefault("application_server.integration.mqtt.location_topic_template", "application/{{ .ApplicationID }}/device/{{ .DevEUI }}/location")
	viper.SetDefault("application_server.integration.mqtt.clean_session", true)
	viper.SetDefault("application_server.integration.enabled", []string{"mqtt"})
	viper.SetDefault("application_server.codec.js.max_execution_time", 100*time.Millisecond)

	viper.SetDefault("application_server.remote_multicast_setup.sync_interval", time.Second)
	viper.SetDefault("application_server.remote_multicast_setup.sync_retries", 3)
	viper.SetDefault("application_server.remote_multicast_setup.sync_batch_size", 100)

	viper.SetDefault("application_server.fragmentation_session.sync_interval", time.Second)
	viper.SetDefault("application_server.fragmentation_session.sync_retries", 3)
	viper.SetDefault("application_server.fragmentation_session.sync_batch_size", 100)

	viper.SetDefault("metrics.timezone", "Local")
	viper.SetDefault("metrics.redis.aggregation_intervals", []string{"MINUTE", "HOUR", "DAY", "MONTH"})
	viper.SetDefault("metrics.redis.minute_aggregation_ttl", time.Hour*2)
	viper.SetDefault("metrics.redis.hour_aggregation_ttl", time.Hour*48)
	viper.SetDefault("metrics.redis.day_aggregation_ttl", time.Hour*24*90)
	viper.SetDefault("metrics.redis.month_aggregation_ttl", time.Hour*24*730)

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

	viperBindEnvs(config.C)

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}

	// backwards compatibility
	if config.C.ApplicationServer.Integration.Backend != "" {
		config.C.ApplicationServer.Integration.Enabled = []string{config.C.ApplicationServer.Integration.Backend}
	}
}

func viperBindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			tv = strings.ToLower(t.Name)
		}
		if tv == "-" {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			viperBindEnvs(v.Interface(), append(parts, tv)...)
		default:
			key := strings.Join(append(parts, tv), ".")
			viper.BindEnv(key)
		}
	}
}
