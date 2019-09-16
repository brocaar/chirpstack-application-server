package config

import (
	"time"

	"github.com/brocaar/lora-app-server/internal/integration/awssns"
	"github.com/brocaar/lora-app-server/internal/integration/azureservicebus"
	"github.com/brocaar/lora-app-server/internal/integration/gcppubsub"
	"github.com/brocaar/lora-app-server/internal/integration/mqtt"
	"github.com/brocaar/lora-app-server/internal/integration/postgresql"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel               int `mapstructure:"log_level"`
		PasswordHashIterations int `mapstructure:"password_hash_iterations"`
	}

	PostgreSQL struct {
		DSN         string `mapstructure:"dsn"`
		Automigrate bool
	} `mapstructure:"postgresql"`

	Redis struct {
		URL         string        `mapstructure:"url"`
		MaxIdle     int           `mapstructure:"max_idle"`
		IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	}

	ApplicationServer struct {
		ID string `mapstructure:"id"`

		Codec struct {
			JS struct {
				MaxExecutionTime time.Duration `mapstructure:"max_execution_time"`
			} `mapstructure:"js"`
		} `mapstructure:"codec"`

		Integration struct {
			Backend         string                 `mapstructure:"backend"` // deprecated
			Enabled         []string               `mapstructure:"enabled"`
			AWSSNS          awssns.Config          `mapstructure:"aws_sns"`
			AzureServiceBus azureservicebus.Config `mapstructure:"azure_service_bus"`
			MQTT            mqtt.Config            `mapstructure:"mqtt"`
			GCPPubSub       gcppubsub.Config       `mapstructure:"gcp_pub_sub"`
			PostgreSQL      postgresql.Config      `mapstructure:"postgresql"`
		}

		API struct {
			Bind       string
			CACert     string `mapstructure:"ca_cert"`
			TLSCert    string `mapstructure:"tls_cert"`
			TLSKey     string `mapstructure:"tls_key"`
			PublicHost string `mapstructure:"public_host"`
		} `mapstructure:"api"`

		ExternalAPI struct {
			Bind                       string
			TLSCert                    string `mapstructure:"tls_cert"`
			TLSKey                     string `mapstructure:"tls_key"`
			JWTSecret                  string `mapstructure:"jwt_secret"`
			DisableAssignExistingUsers bool   `mapstructure:"disable_assign_existing_users"`
			CORSAllowOrigin            string `mapstructure:"cors_allow_origin"`
		} `mapstructure:"external_api"`

		RemoteMulticastSetup struct {
			SyncInterval  time.Duration `mapstructure:"sync_interval"`
			SyncRetries   int           `mapstructure:"sync_retries"`
			SyncBatchSize int           `mapstructure:"sync_batch_size"`
		} `mapstructure:"remote_multicast_setup"`

		FragmentationSession struct {
			SyncInterval  time.Duration `mapstructure:"sync_interval"`
			SyncRetries   int           `mapstructure:"sync_retries"`
			SyncBatchSize int           `mapstructure:"sync_batch_size"`
		} `mapstructure:"fragmentation_session"`

		FUOTADeployment struct {
			McGroupID int `mapstructure:"mc_group_id"`
			FragIndex int `mapstructure:"frag_index"`
		} `mapstructure:"fuota_deployment"`

		Branding struct {
			Header       string
			Footer       string
			Registration string
		}
	} `mapstructure:"application_server"`

	JoinServer struct {
		Bind    string
		CACert  string `mapstructure:"ca_cert"`
		TLSCert string `mapstructure:"tls_cert"`
		TLSKey  string `mapstructure:"tls_key"`

		KEK struct {
			ASKEKLabel string `mapstructure:"as_kek_label"`

			Set []struct {
				Label string `mapstructure:"label"`
				KEK   string `mapstructure:"kek"`
			}
		} `mapstructure:"kek"`
	} `mapstructure:"join_server"`

	Metrics struct {
		Timezone string `mapstructure:"timezone"`
		Redis    struct {
			AggregationIntervals []string      `mapstructure:"aggregation_intervals"`
			MinuteAggregationTTL time.Duration `mapstructure:"minute_aggregation_ttl"`
			HourAggregationTTL   time.Duration `mapstructure:"hour_aggregation_ttl"`
			DayAggregationTTL    time.Duration `mapstructure:"day_aggregation_ttl"`
			MonthAggregationTTL  time.Duration `mapstructure:"month_aggregation_ttl"`
		} `mapstructure:"redis"`
		Prometheus struct {
			EndpointEnabled    bool   `mapstructure:"endpoint_enabled"`
			Bind               string `mapstructure:"bind"`
			APITimingHistogram bool   `mapstructure:"api_timing_histogram"`
		}
	} `mapstructure:"metrics"`
}

// C holds the global configuration.
var C Config
