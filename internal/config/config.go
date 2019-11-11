package config

import (
	"time"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel               int `mapstructure:"log_level"`
		PasswordHashIterations int `mapstructure:"password_hash_iterations"`
	}

	PostgreSQL struct {
		DSN                string `mapstructure:"dsn"`
		Automigrate        bool
		MaxOpenConnections int `mapstructure:"max_open_connections"`
		MaxIdleConnections int `mapstructure:"max_idle_connections"`
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
			Marshaler       string                      `mapstructure:"marshaler"`
			Backend         string                      `mapstructure:"backend"` // deprecated
			Enabled         []string                    `mapstructure:"enabled"`
			AWSSNS          IntegrationAWSSNSConfig     `mapstructure:"aws_sns"`
			AzureServiceBus IntegrationAzureConfig      `mapstructure:"azure_service_bus"`
			MQTT            IntegrationMQTTConfig       `mapstructure:"mqtt"`
			GCPPubSub       IntegrationGCPConfig        `mapstructure:"gcp_pub_sub"`
			PostgreSQL      IntegrationPostgreSQLConfig `mapstructure:"postgresql"`
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

// IntegrationMQTTConfig holds the configuration for the MQTT integration.
type IntegrationMQTTConfig struct {
	Server                  string
	Username                string
	Password                string
	QOS                     uint8  `mapstructure:"qos"`
	CleanSession            bool   `mapstructure:"clean_session"`
	ClientID                string `mapstructure:"client_id"`
	CACert                  string `mapstructure:"ca_cert"`
	TLSCert                 string `mapstructure:"tls_cert"`
	TLSKey                  string `mapstructure:"tls_key"`
	UplinkTopicTemplate     string `mapstructure:"uplink_topic_template"`
	DownlinkTopicTemplate   string `mapstructure:"downlink_topic_template"`
	JoinTopicTemplate       string `mapstructure:"join_topic_template"`
	AckTopicTemplate        string `mapstructure:"ack_topic_template"`
	ErrorTopicTemplate      string `mapstructure:"error_topic_template"`
	StatusTopicTemplate     string `mapstructure:"status_topic_template"`
	LocationTopicTemplate   string `mapstructure:"location_topic_template"`
	UplinkRetainedMessage   bool   `mapstructure:"uplink_retained_message"`
	JoinRetainedMessage     bool   `mapstructure:"join_retained_message"`
	AckRetainedMessage      bool   `mapstructure:"ack_retained_message"`
	ErrorRetainedMessage    bool   `mapstructure:"error_retained_message"`
	StatusRetainedMessage   bool   `mapstructure:"status_retained_message"`
	LocationRetainedMessage bool   `mapstructure:"location_retained_message"`
}

// IntegrationAWSConfig holds the AWS SNS integration configuration.
type IntegrationAWSSNSConfig struct {
	AWSRegion          string `mapstructure:"aws_region"`
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key"`
	TopicARN           string `mapstructure:"topic_arn"`
}

// IntegrationAzureConfig holds the Azure Service-Bus integration configuration.
type IntegrationAzureConfig struct {
	ConnectionString string           `mapstructure:"connection_string"`
	PublishMode      AzurePublishMode `mapstructure:"publish_mode"`
	PublishName      string           `mapstructure:"publish_name"`
}

// IntegrationGCPConfig holds the GCP Pub/Sub integration configuration.
type IntegrationGCPConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	ProjectID       string `mapstructure:"project_id"`
	TopicName       string `mapstructure:"topic_name"`
}

// IntegrationPostgreSQLConfig holds the PostgreSQL integration configuration.
type IntegrationPostgreSQLConfig struct {
	DSN string `json:"dsn"`
}

// AzurePublishMode defines the publish-mode type.
type AzurePublishMode string

// Publish modes.
const (
	AzurePublishModeTopic AzurePublishMode = "topic"
	AzurePublishModeQueue AzurePublishMode = "queue"
)

// C holds the global configuration.
var C Config
