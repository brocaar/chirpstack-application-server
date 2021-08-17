package cmd

import (
	"os"
	"text/template"

	"github.com/brocaar/chirpstack-application-server/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// when updating this template, don't forget to update config.md!
const configTemplate = `[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}

# Log to syslog.
#
# When set to true, log messages are being written to syslog.
log_to_syslog={{ .General.LogToSyslog }}

# The number of times passwords must be hashed. A higher number is safer as
# an attack takes more time to perform.
password_hash_iterations={{ .General.PasswordHashIterations }}

# gRPC default resolver scheme.
#
# Set this to "dns" for enabling dns round-robin load balancing.
grpc_default_resolver_scheme="{{ .General.GRPCDefaultResolverScheme }}"


# PostgreSQL settings.
[postgresql]
# PostgreSQL dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable).
#
# Besides using an URL (e.g. 'postgres://user:password@hostname/database?sslmode=disable')
# it is also possible to use the following format:
# 'user=chirpstack_as dbname=chirpstack_as sslmode=disable'.
#
# The following connection parameters are supported:
#
# * dbname - The name of the database to connect to
# * user - The user to sign in as
# * password - The user's password
# * host - The host to connect to. Values that start with / are for unix domain sockets. (default is localhost)
# * port - The port to bind to. (default is 5432)
# * sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
# * fallback_application_name - An application_name to fall back to if one isn't provided.
# * connect_timeout - Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
# * sslcert - Cert file location. The file must contain PEM encoded data.
# * sslkey - Key file location. The file must contain PEM encoded data.
# * sslrootcert - The location of the root certificate file. The file must contain PEM encoded data.
#
# Valid values for sslmode are:
#
# * disable - No SSL
# * require - Always SSL (skip verification)
# * verify-ca - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
# * verify-full - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)
dsn="{{ .PostgreSQL.DSN }}"

# Automatically apply database migrations.
#
# It is possible to apply the database-migrations by hand
# (see https://github.com/brocaar/chirpstack-application-server/tree/master/internal/storage/migrations)
# or let ChirpStack Application Server migrate to the latest state automatically, by using
# this setting. Make sure that you always make a backup when upgrading Lora
# App Server and / or applying migrations.
automigrate={{ .PostgreSQL.Automigrate }}

# Max open connections.
#
# This sets the max. number of open connections that are allowed in the
# PostgreSQL connection pool (0 = unlimited).
max_open_connections={{ .PostgreSQL.MaxOpenConnections }}

# Max idle connections.
#
# This sets the max. number of idle connections in the PostgreSQL connection
# pool (0 = no idle connections are retained).
max_idle_connections={{ .PostgreSQL.MaxIdleConnections }}


# Redis settings
[redis]

# Server address or addresses.
#
# Set multiple addresses when connecting to a cluster.
servers=[{{ range $index, $elm := .Redis.Servers }}
  "{{ $elm }}",{{ end }}
]

# Password.
#
# Set the password when connecting to Redis requires password authentication.
password="{{ .Redis.Password }}"

# Database index.
#
# By default, this can be a number between 0-15.
database={{ .Redis.Database }}


# Redis Cluster.
#
# Set this to true when the provided URLs are pointing to a Redis Cluster
# instance.
cluster={{ .Redis.Cluster }}

# Master name.
#
# Set the master name when the provided URLs are pointing to a Redis Sentinel
# instance.
master_name="{{ .Redis.MasterName }}"

# Connection pool size.
#
# Default (when set to 0) is 10 connections per every CPU.
pool_size={{ .Redis.PoolSize }}

# TLS enabled.
#
# Note: tis will enable TLS, but it will not validate the certificate
# used by the server.
tls_enabled={{ .Redis.TLSEnabled }}

# Key prefix.
#
# A key prefix can be used to avoid key collisions when multiple deployments
# are using the same Redis database and it is not possible to separate
# keys by database index (e.g. when using Redis Cluster, which does not
# support multiple databases).
key_prefix="{{ .Redis.KeyPrefix }}"


# Application-server settings.
[application_server]
# Application-server identifier.
#
# Random UUID defining the id of the application-server installation (used by
# ChirpStack Network Server as routing-profile id).
# For now it is recommended to not change this id.
id="{{ .ApplicationServer.ID }}"


  # User authentication
  [application_server.user_authentication]

    # OpenID Connect.
    [application_server.user_authentication.openid_connect]

    # Enable OpenID Connect authentication.
    #
    # Enabling this option replaces password authentication.
    enabled={{ .ApplicationServer.UserAuthentication.OpenIDConnect.Enabled }}

    # Registration enabled.
    #
    # Enabling this will automatically register the user when it is not yet present
    # in the ChirpStack Application Server database. There is no
    # registration form as the user information is automatically received using the
    # OpenID Connect provided information.
    # The user will not be associated with any organization, but in order to
    # facilitate the automatic onboarding of users, it is possible to configure a
    # registration callback URL (next config option).
    registration_enabled={{ .ApplicationServer.UserAuthentication.OpenIDConnect.RegistrationEnabled }}

    # Registration callback URL.
    #
    # This (optional) endpoint will be called on the registration of the user and
    # can implement the association of the user with an organization, create a new
    # organization, ...
    # ChirpStack Application Server will make a HTTP POST call to this endpoint,
    # with the following URL parameters:
    # - user_id, of the newly created user in ChirpStack Application Server.
    # - oidc_claims, the claims returned by the OpenID Connect "UserInfo" call, JSON-encoded. Use this to find additional information, like the users organization.
    registration_callback_url="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.RegistrationCallbackURL }}"

    # Provider URL.
    # This is the URL of the OpenID Connect provider.
    # Example: https://auth.example.org
    provider_url="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.ProviderURL }}"

    # Client ID.
    client_id="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.ClientID }}"

    # Client secret.
    client_secret="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.ClientSecret }}"

    # Redirect URL.
    #
    # This must contain the ChirpStack Application Server web-interface hostname
    # with '/auth/oidc/callback' path, e.g. https://example.com/auth/oidc/callback.
    redirect_url="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.RedirectURL }}"

    # Logout URL.
    #
    # When set, ChirpStack Application Server will redirect to this URL instead
    # of redirecting to the login page.
    logout_url="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.LogoutURL }}"

    # Login label.
    #
    # The login label is used in the web-interface login form.
    login_label="{{ .ApplicationServer.UserAuthentication.OpenIDConnect.LoginLabel }}"


  # JavaScript codec settings.
  [application_server.codec.js]
  # Maximum execution time.
  max_execution_time="{{ .ApplicationServer.Codec.JS.MaxExecutionTime }}"


  # Integration configures the data integration.
  #
  # This is the data integration which is available for all applications,
  # besides the extra integrations that can be added on a per-application
  # basis.
  [application_server.integration]
  # Payload marshaler.
  #
  # This defines how the MQTT payloads are encoded. Valid options are:
  # * protobuf:  Protobuf encoding
  # * json:      JSON encoding (easier for debugging, but less compact than 'protobuf')
  # * json_v3:   v3 JSON (will be removed in the next major release)
  marshaler="{{ .ApplicationServer.Integration.Marshaler }}"


  # Enabled integrations.
  #
  # Enabled integrations are enabled for all applications. Multiple
  # integrations can be configured.
  # Do not forget to configure the related configuration section below for
  # the enabled integrations. Integrations that can be enabled are:
  # * mqtt              - MQTT broker
  # * amqp              - AMQP / RabbitMQ
  # * aws_sns           - AWS Simple Notification Service (SNS)
  # * azure_service_bus - Azure Service-Bus
  # * gcp_pub_sub       - Google Cloud Pub/Sub
  # * kafka             - Kafka distributed streaming platform
  # * postgresql        - PostgreSQL database
  enabled=[{{ if .ApplicationServer.Integration.Enabled|len }}"{{ end }}{{ range $index, $elm := .ApplicationServer.Integration.Enabled }}{{ if $index }}", "{{ end }}{{ $elm }}{{ end }}{{ if .ApplicationServer.Integration.Enabled|len }}"{{ end }}]


  # MQTT integration backend.
  [application_server.integration.mqtt]
  # Event topic template.
  event_topic_template="{{ .ApplicationServer.Integration.MQTT.EventTopicTemplate }}"

  # Command topic template.
  command_topic_template="{{ .ApplicationServer.Integration.MQTT.CommandTopicTemplate }}"

  # Retain events.
  #
  # The MQTT broker will store the last publised message, when retain events is set
  # to true. When a MQTT client connects and subscribes, it will always receive the
  # last published message.
  retain_events={{ .ApplicationServer.Integration.MQTT.RetainEvents }}

  # MQTT server (e.g. scheme://host:port where scheme is tcp, ssl or ws)
  server="{{ .ApplicationServer.Integration.MQTT.Server }}"

  # Connect with the given username (optional)
  username="{{ .ApplicationServer.Integration.MQTT.Username }}"

  # Connect with the given password (optional)
  password="{{ .ApplicationServer.Integration.MQTT.Password }}"

  # Maximum interval that will be waited between reconnection attempts when connection is lost.
  # Valid units are 'ms', 's', 'm', 'h'. Note that these values can be combined, e.g. '24h30m15s'.
  max_reconnect_interval="{{ .ApplicationServer.Integration.MQTT.MaxReconnectInterval }}"

  # Quality of service level
  #
  # 0: at most once
  # 1: at least once
  # 2: exactly once
  #
  # Note: an increase of this value will decrease the performance.
  # For more information: https://www.hivemq.com/blog/mqtt-essentials-part-6-mqtt-quality-of-service-levels
  qos={{ .ApplicationServer.Integration.MQTT.QOS }}

  # Clean session
  #
  # Set the "clean session" flag in the connect message when this client
  # connects to an MQTT broker. By setting this flag you are indicating
  # that no messages saved by the broker for this client should be delivered.
  clean_session={{ .ApplicationServer.Integration.MQTT.CleanSession }}

  # Client ID
  #
  # Set the client id to be used by this client when connecting to the MQTT
  # broker. A client id must be no longer than 23 characters. When left blank,
  # a random id will be generated. This requires clean_session=true.
  client_id="{{ .ApplicationServer.Integration.MQTT.ClientID }}"

  # CA certificate file (optional)
  #
  # Use this when setting up a secure connection (when server uses ssl://...)
  # but the certificate used by the server is not trusted by any CA certificate
  # on the server (e.g. when self generated).
  ca_cert="{{ .ApplicationServer.Integration.MQTT.CACert }}"

  # TLS certificate file (optional)
  tls_cert="{{ .ApplicationServer.Integration.MQTT.TLSCert }}"

  # TLS key file (optional)
  tls_key="{{ .ApplicationServer.Integration.MQTT.TLSKey }}"

    # Client configuration.
    #
    # This section contains configuration for end-applications connecting to the
    # MQTT broker.
    [application_server.integration.mqtt.client]

    # CA certificate and key file (optional).
    #
    # When setting the CA certificate and key file options, ChirpStack Application Server
    # will generate client certificates which can be used by the end-application for
    # authentication and authorization with the MQTT broker. The Common Name of the
    # certificate will be set to the application ID.
    ca_cert="{{ .ApplicationServer.Integration.MQTT.Client.CACert }}"
    ca_key="{{ .ApplicationServer.Integration.MQTT.Client.CAKey }}"
    client_cert_lifetime="{{ .ApplicationServer.Integration.MQTT.Client.ClientCertLifetime }}"


  # AMQP / RabbitMQ.
  [application_server.integration.amqp]
  # Server URL.
  #
  # See for a specification of all the possible options:
  # https://www.rabbitmq.com/uri-spec.html
  url="{{ .ApplicationServer.Integration.AMQP.URL }}"

  # Event routing key template.
  #
  # This is the event routing-key template used when publishing device
  # events. Messages will be published to the "amq.topic" exchange.
  event_routing_key_template="{{ .ApplicationServer.Integration.AMQP.EventRoutingKeyTemplate }}"


  # AWS Simple Notification Service (SNS)
  [application_server.integration.aws_sns]
  # AWS region.
  #
  # Example: "eu-west-1".
  # See also: https://docs.aws.amazon.com/general/latest/gr/rande.html.
  aws_region="{{ .ApplicationServer.Integration.AWSSNS.AWSRegion }}"

  # AWS Access Key ID.
  aws_access_key_id="{{ .ApplicationServer.Integration.AWSSNS.AWSAccessKeyID }}"

  # AWS Secret Access Key.
  aws_secret_access_key="{{ .ApplicationServer.Integration.AWSSNS.AWSSecretAccessKey }}"

  # Topic ARN (SNS).
  topic_arn="{{ .ApplicationServer.Integration.AWSSNS.TopicARN }}"


  # Azure Service-Bus integration.
  [application_server.integration.azure_service_bus]
  # Connection string.
  #
  # The connection string can be found / created in the Azure console under
  # Settings -> Shared access policies. The policy must contain Manage & Send.
  connection_string="{{ .ApplicationServer.Integration.AzureServiceBus.ConnectionString }}"

  # Publish mode.
  #
  # Select either "topic", or "queue".
  publish_mode="{{ .ApplicationServer.Integration.AzureServiceBus.PublishMode }}"

  # Publish name.
  #
  # The name of the topic or queue.
  publish_name="{{ .ApplicationServer.Integration.AzureServiceBus.PublishName }}"


  # Google Cloud Pub/Sub integration.
  [application_server.integration.gcp_pub_sub]
  # Path to the IAM service-account credentials file.
  #
  # Note: this service-account must have the following Pub/Sub roles:
  #  * Pub/Sub Editor
  credentials_file="{{ .ApplicationServer.Integration.GCPPubSub.CredentialsFile }}"

  # Google Cloud project id.
  project_id="{{ .ApplicationServer.Integration.GCPPubSub.ProjectID }}"

  # Pub/Sub topic name.
  topic_name="{{ .ApplicationServer.Integration.GCPPubSub.TopicName }}"


  # Kafka integration.
  [application_server.integration.kafka]
  # Brokers, e.g.: localhost:9092.
  brokers=[{{ range $index, $broker := .ApplicationServer.Integration.Kafka.Brokers }}{{ if $index }}, {{ end }}"{{ $broker }}"{{ end }}]

  # TLS.
  #
  # Set this to true when the Kafka client must connect using TLS to the Broker.
  tls={{ .ApplicationServer.Integration.Kafka.TLS }}

  # Topic for events.
  topic="{{ .ApplicationServer.Integration.Kafka.Topic }}"

  # Template for keys included in Kafka messages. If empty, no key is included.
  # Kafka uses the key for distributing messages over partitions. You can use
  # this to ensure some subset of messages end up in the same partition, so
  # they can be consumed in-order. And Kafka can use the key for data retention
  # decisions.  A header "event" with the event type is included in each
  # message. There is no need to parse it from the key.
  event_key_template="{{ .ApplicationServer.Integration.Kafka.EventKeyTemplate }}"

  # Username (optional).
  username="{{ .ApplicationServer.Integration.Kafka.Username }}"

  # Password (optional).
  password="{{ .ApplicationServer.Integration.Kafka.Password }}"

  # One of plain or scram
  mechanism="{{ .ApplicationServer.Integration.Kafka.Mechanism }}"
  
  # Only used if mechanism == scram.
  # SHA-256 or SHA-512 
  algorithm="{{ .ApplicationServer.Integration.Kafka.Algorithm }}"

  # PostgreSQL database integration.
  [application_server.integration.postgresql]
  # PostgreSQL dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable).
  dsn="{{ .ApplicationServer.Integration.PostgreSQL.DSN }}"

  # This sets the max. number of open connections that are allowed in the
  # PostgreSQL connection pool (0 = unlimited).
  max_open_connections={{ .ApplicationServer.Integration.PostgreSQL.MaxOpenConnections }}

  # Max idle connections.
  #
  # This sets the max. number of idle connections in the PostgreSQL connection
  # pool (0 = no idle connections are retained).
  max_idle_connections={{ .ApplicationServer.Integration.PostgreSQL.MaxIdleConnections }}


  # Settings for the "internal api"
  #
  # This is the API used by ChirpStack Network Server to communicate with ChirpStack Application Server
  # and should not be exposed to the end-user.
  [application_server.api]
  # ip:port to bind the api server
  bind="{{ .ApplicationServer.API.Bind }}"

  # ca certificate used by the api server (optional)
  ca_cert="{{ .ApplicationServer.API.CACert }}"

  # tls certificate used by the api server (optional)
  tls_cert="{{ .ApplicationServer.API.TLSCert }}"

  # tls key used by the api server (optional)
  tls_key="{{ .ApplicationServer.API.TLSKey }}"

  # Public ip:port of the application-server API.
  #
  # This is used by ChirpStack Network Server to connect to ChirpStack Application Server. When running
  # ChirpStack Application Server on a different host than ChirpStack Network Server, make sure to set
  # this to the host:ip on which ChirpStack Network Server can reach ChirpStack Application Server.
  # The port must be equal to the port configured by the 'bind' flag
  # above.
  public_host="{{ .ApplicationServer.API.PublicHost }}"


  # Settings for the "external api"
  #
  # This is the API and web-interface exposed to the end-user.
  [application_server.external_api]
  # ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api)
  bind="{{ .ApplicationServer.ExternalAPI.Bind }}"

  # http server TLS certificate (optional)
  tls_cert="{{ .ApplicationServer.ExternalAPI.TLSCert }}"

  # http server TLS key (optional)
  tls_key="{{ .ApplicationServer.ExternalAPI.TLSKey }}"

  # JWT secret used for api authentication / authorization
  # You could generate this by executing 'openssl rand -base64 32' for example
  jwt_secret="{{ .ApplicationServer.ExternalAPI.JWTSecret }}"

  # Allow origin header (CORS).
  #
  # Set this to allows cross-domain communication from the browser (CORS).
  # Example value: https://example.com.
  # When left blank (default), CORS will not be used.
  cors_allow_origin="{{ .ApplicationServer.ExternalAPI.CORSAllowOrigin }}"


{{ if ne .ApplicationServer.Branding.Footer  "" }}
  # Branding configuration.
  [application_server.branding]
  # Footer
  footer="{{ .ApplicationServer.Branding.Footer }}"

  # Registration.
  registration="{{ .ApplicationServer.Branding.Registration }}"

{{ end }}

# Join-server configuration.
#
# ChirpStack Application Server implements a (subset) of the join-api specified by the
# LoRaWAN Backend Interfaces specification. This API is used by ChirpStack Network Server
# to handle join-requests.
[join_server]
# ip:port to bind the join-server api interface to
bind="{{ .JoinServer.Bind }}"

# CA certificate (optional).
#
# When set, the server requires a client-certificate and will validate this
# certificate on incoming requests.
ca_cert="{{ .JoinServer.CACert }}"

# TLS server-certificate (optional).
#
# Set this to enable TLS.
tls_cert="{{ .JoinServer.TLSCert }}"

# TLS server-certificate key (optional).
#
# Set this to enable TLS.
tls_key="{{ .JoinServer.TLSKey }}"


# Key Encryption Key (KEK) configuration.
#
# The KEK mechanism is used to encrypt the session-keys sent from the
# join-server to the network-server.
#
# The ChirpStack Application Server join-server will use the NetID of the requesting
# network-server as the KEK label. When no such label exists in the set,
# the session-keys will be sent unencrypted (which can be fine for
# private networks).
#
# Please refer to the LoRaWAN Backend Interface specification
# 'Key Transport Security' section for more information.
[join_server.kek]

  # Application-server KEK label.
  #
  # This defines the KEK label used to encrypt the AppSKey (note that the
  # AppSKey is signaled to the NS and on the first received uplink from the
  # NS to the AS).
  #
  # When left blank, the AppSKey will be sent unencrypted (which can be fine
  # for private networks).
  as_kek_label="{{ .JoinServer.KEK.ASKEKLabel }}"

  # KEK set.
  #
  # Example (the [[join_server.kek.set]] can be repeated):
  # [[join_server.kek.set]]
  # # KEK label.
  # label="000000"

  # # Key Encryption Key.
  # kek="01020304050607080102030405060708"
{{ range $index, $element := .JoinServer.KEK.Set }}
  [[join_server.kek.set]]
  label="{{ $element.Label }}"
  kek="{{ $element.KEK }}"
{{ end }}

# Metrics collection settings.
[metrics]
# Timezone
#
# The timezone is used for correctly aggregating the metrics (e.g. per hour,
# day or month).
# Example: "Europe/Amsterdam" or "Local" for the the system's local time zone.
timezone="{{ .Metrics.Timezone }}"

  # Metrics stored in Redis.
  #
  # The following metrics are stored in Redis:
  # * gateway statistics
  [metrics.redis]
  # Aggregation intervals
  #
  # The intervals on which to aggregate. Available options are:
  # 'MINUTE', 'HOUR', 'DAY', 'MONTH'.
  aggregation_intervals=[{{ if .Metrics.Redis.AggregationIntervals|len }}"{{ end }}{{ range $index, $elm := .Metrics.Redis.AggregationIntervals }}{{ if $index }}", "{{ end }}{{ $elm }}{{ end }}{{ if .Metrics.Redis.AggregationIntervals|len }}"{{ end }}]

  # Aggregated statistics storage duration.
  minute_aggregation_ttl="{{ .Metrics.Redis.MinuteAggregationTTL }}"
  hour_aggregation_ttl="{{ .Metrics.Redis.HourAggregationTTL }}"
  day_aggregation_ttl="{{ .Metrics.Redis.DayAggregationTTL }}"
  month_aggregation_ttl="{{ .Metrics.Redis.MonthAggregationTTL }}"


  # Metrics stored in Prometheus.
  #
  # These metrics expose information about the state of the ChirpStack Network Server
  # instance.
  [metrics.prometheus]
  # Enable Prometheus metrics endpoint.
  endpoint_enabled={{ .Metrics.Prometheus.EndpointEnabled }}

  # The ip:port to bind the Prometheus metrics server to for serving the
  # metrics endpoint.
  bind="{{ .Metrics.Prometheus.Bind }}"

  # API timing histogram.
  #
  # By setting this to true, the API request timing histogram will be enabled.
  # See also: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
  api_timing_histogram={{ .Metrics.Prometheus.APITimingHistogram }}


  # Monitoring settings.
  #
  # Note that this replaces the metrics.prometheus configuration. If a
  # metrics.prometheus if found in the configuration then it will fall back
  # to that and the monitoring section is ignored.
  [monitoring]

  # IP:port to bind the monitoring endpoint to.
  #
  # When left blank, the monitoring endpoint will be disabled.
  bind="{{ .Monitoring.Bind }}"

  # Prometheus metrics endpoint.
  #
  # When set to true, Prometheus metrics will be served at '/metrics'.
  prometheus_endpoint={{ .Monitoring.PrometheusEndpoint }}

  # Prometheus API timing histogram.
  #
  # By setting this to true, the API request timing histogram will be enabled.
  # See also: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
  prometheus_api_timing_histogram={{ .Monitoring.PrometheusAPITimingHistogram }}

  # Health check endpoint.
  #
  # When set to true, the healthcheck endpoint will be served at '/health'.
  # When requesting, this endpoint will perform the following actions to
  # determine the health of this service:
  #   * Ping PostgreSQL database
  #   * Ping Redis database
  healthcheck_endpoint={{ .Monitoring.HealthcheckEndpoint }}

  # Per device event-log max history.
  #
  # When set to > 0, ChirpStack Application Server will log events per device
  # to a Redis Stream. This feature is used by the web-interface.
  per_device_event_log_max_history={{ .Monitoring.PerDeviceEventLogMaxHistory }}

`

var configCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print the LoRa Application Server configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		err := t.Execute(os.Stdout, config.C)
		if err != nil {
			return errors.Wrap(err, "execute config template error")
		}
		return nil
	},
}
