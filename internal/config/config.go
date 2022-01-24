package config

import (
	"time"
)

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel                  int    `mapstructure:"log_level"`
		LogToSyslog               bool   `mapstructure:"log_to_syslog"`
		PasswordHashIterations    int    `mapstructure:"password_hash_iterations"`
		GRPCDefaultResolverScheme string `mapstructure:"grpc_default_resolver_scheme"`
	} `mapstructure:"general"`

	PostgreSQL struct {
		DSN                string `mapstructure:"dsn"`
		Automigrate        bool
		MaxOpenConnections int `mapstructure:"max_open_connections"`
		MaxIdleConnections int `mapstructure:"max_idle_connections"`
	} `mapstructure:"postgresql"`

	Redis struct {
		URL        string   `mapstructure:"url"` // deprecated
		Servers    []string `mapstructure:"servers"`
		Cluster    bool     `mapstructure:"cluster"`
		MasterName string   `mapstructure:"master_name"`
		PoolSize   int      `mapstructure:"pool_size"`
		Password   string   `mapstructure:"password"`
		Database   int      `mapstructure:"database"`
		TLSEnabled bool     `mapstructure:"tls_enabled"`
		KeyPrefix  string   `mapstructure:"key_prefix"`
	} `mapstructure:"redis"`

	ApplicationServer struct {
		ID string `mapstructure:"id"`

		UserAuthentication struct {
			OpenIDConnect struct {
				Enabled                 bool   `mapstructure:"enabled"`
				RegistrationEnabled     bool   `mapstructure:"registration_enabled"`
				RegistrationCallbackURL string `mapstructure:"registration_callback_url"`
				ProviderURL             string `mapstructure:"provider_url"`
				ClientID                string `mapstructure:"client_id"`
				ClientSecret            string `mapstructure:"client_secret"`
				RedirectURL             string `mapstructure:"redirect_url"`
				LogoutURL               string `mapstructure:"logout_url"`
				LoginLabel              string `mapstructure:"login_label"`
			} `mapstructure:"openid_connect"`
		} `mapstructure:"user_authentication"`

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
			Kafka           IntegrationKafkaConfig      `mapstructure:"kafka"`
			PostgreSQL      IntegrationPostgreSQLConfig `mapstructure:"postgresql"`
			AMQP            IntegrationAMQPConfig       `mapstructure:"amqp"`
		} `mapstructure:"integration"`

		API struct {
			Bind       string
			CACert     string `mapstructure:"ca_cert"`
			TLSCert    string `mapstructure:"tls_cert"`
			TLSKey     string `mapstructure:"tls_key"`
			PublicHost string `mapstructure:"public_host"`
		} `mapstructure:"api"`

		ExternalAPI struct {
			Bind            string
			TLSCert         string `mapstructure:"tls_cert"`
			TLSKey          string `mapstructure:"tls_key"`
			JWTSecret       string `mapstructure:"jwt_secret"`
			CORSAllowOrigin string `mapstructure:"cors_allow_origin"`
		} `mapstructure:"external_api"`

		Branding struct {
			Footer       string
			Registration string
		} `mapstructure:"branding"`
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
		} `mapstructure:"prometheus"`
	} `mapstructure:"metrics"`

	Monitoring struct {
		Bind                         string `mapstructure:"bind"`
		PrometheusEndpoint           bool   `mapstructure:"prometheus_endpoint"`
		PrometheusAPITimingHistogram bool   `mapstructure:"prometheus_api_timing_histogram"`
		HealthcheckEndpoint          bool   `mapstructure:"healthcheck_endpoint"`
		PerDeviceEventLogMaxHistory  int64  `mapstructure:"per_device_event_log_max_history"`
	} `mapstructure:"monitoring"`
}

// IntegrationMQTTConfig holds the configuration for the MQTT integration.
type IntegrationMQTTConfig struct {
	Client IntegrationMQTTClientConfig `mapstructure:"client"`

	Server               string        `mapstructure:"server"`
	Username             string        `mapstructure:"username"`
	Password             string        `mapstructure:"password"`
	MaxReconnectInterval time.Duration `mapstructure:"max_reconnect_interval"`
	QOS                  uint8         `mapstructure:"qos"`
	CleanSession         bool          `mapstructure:"clean_session"`
	ClientID             string        `mapstructure:"client_id"`
	CACert               string        `mapstructure:"ca_cert"`
	TLSCert              string        `mapstructure:"tls_cert"`
	TLSKey               string        `mapstructure:"tls_key"`
	EventTopicTemplate   string        `mapstructure:"event_topic_template"`
	CommandTopicTemplate string        `mapstructure:"command_topic_template"`
	RetainEvents         bool          `mapstructure:"retain_events"`

	// For backards compatibility
	UplinkTopicTemplate        string `mapstructure:"uplink_topic_template"`
	DownlinkTopicTemplate      string `mapstructure:"downlink_topic_template"`
	JoinTopicTemplate          string `mapstructure:"join_topic_template"`
	AckTopicTemplate           string `mapstructure:"ack_topic_template"`
	ErrorTopicTemplate         string `mapstructure:"error_topic_template"`
	StatusTopicTemplate        string `mapstructure:"status_topic_template"`
	LocationTopicTemplate      string `mapstructure:"location_topic_template"`
	TxAckTopicTemplate         string `mapstructure:"tx_ack_topic_template"`
	IntegrationTopicTemplate   string `mapstructure:"integration_topic_template"`
	UplinkRetainedMessage      bool   `mapstructure:"uplink_retained_message"`
	JoinRetainedMessage        bool   `mapstructure:"join_retained_message"`
	AckRetainedMessage         bool   `mapstructure:"ack_retained_message"`
	ErrorRetainedMessage       bool   `mapstructure:"error_retained_message"`
	StatusRetainedMessage      bool   `mapstructure:"status_retained_message"`
	LocationRetainedMessage    bool   `mapstructure:"location_retained_message"`
	TxAckRetainedMessage       bool   `mapstructure:"tx_ack_retained_message"`
	IntegrationRetainedMessage bool   `mapstructure:"integration_retained_message"`
}

// IntegrationMQTTClientConfig holds the additional client config.
type IntegrationMQTTClientConfig struct {
	CACert             string        `mapstructure:"ca_cert"`
	CAKey              string        `mapstructure:"ca_key"`
	ClientCertLifetime time.Duration `mapstructure:"client_cert_lifetime"`
}

// IntegrationAWSSNSConfig holds the AWS SNS integration configuration.
type IntegrationAWSSNSConfig struct {
	Marshaler          string `mapstructure:"marshaler" json:"marshaler"`
	AWSRegion          string `mapstructure:"aws_region" json:"region"`
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id" json:"accessKeyID"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key" json:"secretAccessKey"`
	TopicARN           string `mapstructure:"topic_arn" json:"topicARN"`
}

// IntegrationAzureConfig holds the Azure Service-Bus integration configuration.
type IntegrationAzureConfig struct {
	Marshaler        string           `mapstructure:"marshaler" json:"marshaler"`
	ConnectionString string           `mapstructure:"connection_string" json:"connectionString"`
	PublishMode      AzurePublishMode `mapstructure:"publish_mode" json:"-"`
	PublishName      string           `mapstructure:"publish_name" json:"publishName"`
}

// IntegrationGCPConfig holds the GCP Pub/Sub integration configuration.
type IntegrationGCPConfig struct {
	Marshaler            string `mapstructure:"marshaler" json:"marshaler"`
	CredentialsFile      string `mapstructure:"credentials_file" json:"-"`
	CredentialsFileBytes []byte `mapstructure:"-" json:"credentialsFile"`
	ProjectID            string `mapstructure:"project_id" json:"projectID"`
	TopicName            string `mapstructure:"topic_name" json:"topicName"`
}

// IntegrationPostgreSQLConfig holds the PostgreSQL integration configuration.
type IntegrationPostgreSQLConfig struct {
	DSN                string `json:"dsn"`
	MaxOpenConnections int    `mapstructure:"max_open_connections"`
	MaxIdleConnections int    `mapstructure:"max_idle_connections"`
}

// IntegrationAMQPConfig holds the AMQP integration configuration.
type IntegrationAMQPConfig struct {
	URL                     string `mapstructure:"url"`
	EventRoutingKeyTemplate string `mapstructure:"event_routing_key_template"`
}

// IntegrationKafkaConfig holds the Kafka integration configuration.
type IntegrationKafkaConfig struct {
	Brokers          []string `mapstructure:"brokers"`
	TLS              bool     `mapstructure:"tls"`
	Topic            string   `mapstructure:"topic"`
	EventKeyTemplate string   `mapstructure:"event_key_template"`
	Username         string   `mapstructure:"username"`
	Password         string   `mapstructure:"password"`
	Mechanism        string   `mapstructure:"mechanism"`
	Algorithm        string   `mapstructure:"algorithm"`
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

// Get returns the configuration.
func Get() *Config {
	return &C
}

// Set sets the configuration.
func Set(c Config) {
	C = c
}
