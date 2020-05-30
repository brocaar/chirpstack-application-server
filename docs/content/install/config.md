---
title: Configuration
menu:
    main:
        parent: install
        weight: 4
toc: false
description: Instructions and examples how to configure the ChirpStack Application Server service.
---

# Configuration

The `chirpstack-application-server` binary has the following command-line flags:

{{<highlight text>}}
ChirpStack Application Server is an open-source application-server, part of the ChirpStack Network Server project
	> documentation & support: https://www.chirpstack.io/application-server/
	> source & copyright information: https://github.com/brocaar/chirpstack-application-server

Usage:
  chirpstack-application-server [flags]
  chirpstack-application-server [command]

Available Commands:
  configfile  Print the LoRa Application Server configuration file
  help        Help about any command
  version     Print the ChirpStack Application Server version

Flags:
  -c, --config string   path to configuration file (optional)
  -h, --help            help for chirpstack-application-server
      --log-level int   debug=5, info=4, error=2, fatal=1, panic=0 (default 4)

Use "chirpstack-application-server [command] --help" for more information about a command.
{{< /highlight >}}

## Configuration file

By default `chirpstack-application-server` will look in the following order for a
configuration file at the following paths when `--config` is not set:

* `chirpstack-application-server.toml` (current working directory)
* `$HOME/.config/chirpstack-application-server/chirpstack-application-server.toml`
* `/etc/chirpstack-application-server/chirpstack-application-server.toml`

To load configuration from a different location, use the `--config` flag.

To generate a new configuration file `chirpstack-application-server.toml`, execute the following command:

{{<highlight bash>}}
chirpstack-application-server configfile > chirpstack-application-server.toml
{{< /highlight >}}

Note that this configuration file will be pre-filled with the current configuration
(either loaded from the paths mentioned above, or by using the `--config` flag).
This makes it possible when new fields get added to upgrade your configuration file
while preserving your old configuration. Example:

{{<highlight bash>}}
chirpstack-application-server configfile --config chirpstack-application-server-old.toml > chirpstack-application-server-new.toml
{{< /highlight >}}

Example configuration file:

{{<highlight toml>}}
[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level=4

# Log to syslog.
#
# When set to true, log messages are being written to syslog.
log_to_syslog=false

# The number of times passwords must be hashed. A higher number is safer as
# an attack takes more time to perform.
password_hash_iterations=100000


# PostgreSQL settings.
#
# Please note that PostgreSQL 9.5+ is required.
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
dsn="postgres://localhost/chirpstack_as?sslmode=disable"

# Automatically apply database migrations.
#
# It is possible to apply the database-migrations by hand
# (see https://github.com/brocaar/chirpstack-application-server/tree/master/migrations)
# or let ChirpStack Application Server migrate to the latest state automatically, by using
# this setting. Make sure that you always make a backup when upgrading Lora
# App Server and / or applying migrations.
automigrate=true

# Max open connections.
#
# This sets the max. number of open connections that are allowed in the
# PostgreSQL connection pool (0 = unlimited).
max_open_connections=0

# Max idle connections.
#
# This sets the max. number of idle connections in the PostgreSQL connection
# pool (0 = no idle connections are retained).
max_idle_connections=2


# Redis settings
#
# Please note that Redis 2.6.0+ is required.
[redis]
# Redis url (e.g. redis://user:password@hostname/0)
#
# For more information about the Redis URL format, see:
# https://www.iana.org/assignments/uri-schemes/prov/redis
url="redis://localhost:6379"

# Redis Cluster.
#
# Set this to true when the provided URL is pointing to a Redis Cluster
# instance.
cluster=false

# The master name.
#
# Set the master name when the provided URL is pointing to a Redis Sentinel
# instance.
master_name=""

# Connection pool size.
#
# Default is 10 connections per every CPU.
pool_size=0


# Application-server settings.
[application_server]
# Application-server identifier.
#
# Random UUID defining the id of the application-server installation (used by
# ChirpStack Network Server as routing-profile id).
# For now it is recommended to not change this id.
id="6d5db27e-4ce2-4b2b-b5d7-91f069397978"


  # User authentication
  [application_server.user_authentication]

    # OpenID Connect.
    [application_server.user_authentication.openid_connect]

    # Enable OpenID Connect authentication.
    #
    # Enabling this option replaces password authentication.
    enabled=false

    # Registration enabled.
    #
    # Enabling this will automatically register the user when it is not yet present
    # in the ChirpStack Application Server database. There is no
    # registration form as the user information is automatically received using the
    # OpenID Connect provided information.
    # The user will not be associated with any organization, but in order to
    # facilitate the automatic onboarding of users, it is possible to configure a
    # registration callback URL (next config option).
    registration_enabled=false

    # Registration callback URL.
    #
    # This (optional) endpoint will be called on the registration of the user and
    # can implement the association of the user with an organization, create a new
    # organization, ...
    # ChirpStack Application Server will make a HTTP POST call to this endpoint,
    # with the URL parameter user_id.
    registration_callback_url=""

    # Provider URL.
    # This is the URL of the OpenID Connect provider.
    provider_url=""

    # Client ID.
    client_id=""

    # Client secret.
    client_secret=""

    # Redirect URL.
    #
    # This must contain the ChirpStack Application Server web-interface hostname
    # with '/auth/oidc/callback' path, e.g. https://example.com/auth/oidc/callback.
    redirect_url=""

    # Login label.
    #
    # The login label is used in the web-interface login form.
    login_label=""


  # JavaScript codec settings.
  [application_server.codec.js]
  # Maximum execution time.
  max_execution_time="100ms"


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
  marshaler="json_v3"


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
  # * postgresql        - PostgreSQL database
  enabled=["mqtt"]


  # MQTT integration backend.
  [application_server.integration.mqtt]
  # MQTT topic templates for the different MQTT topics.
  #
  # The meaning of these topics are documented at:
  # https://www.chirpstack.io/application-server/integrate/data/
  #
  # The following substitutions can be used:
  # * "{{ .ApplicationID }}" for the application id.
  # * "{{ .DevEUI }}" for the DevEUI of the device.
  #
  # Note: the downlink_topic_template must contain both the application id and
  # DevEUI substitution!
  uplink_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/rx"
  downlink_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/tx"
  join_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/join"
  ack_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/ack"
  error_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/error"
  status_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/status"
  location_topic_template="application/{{ .ApplicationID }}/device/{{ .DevEUI }}/location"

  # Retained messages configuration.
  #
  # The MQTT broker will store the last publised message, when retained message is set
  # to true. When a client subscribes to a topic with retained message set to true, it will
  # always receive the last published message.
  uplink_retained_message=false
  join_retained_message=false
  ack_retained_message=false
  error_retained_message=false
  status_retained_message=false
  location_retained_message=false

  # MQTT server (e.g. scheme://host:port where scheme is tcp, ssl or ws)
  server="tcp://localhost:1883"

  # Connect with the given username (optional)
  username=""

  # Connect with the given password (optional)
  password=""

  # Maximum interval that will be waited between reconnection attempts when connection is lost.
  # Valid units are 'ms', 's', 'm', 'h'. Note that these values can be combined, e.g. '24h30m15s'.
  max_reconnect_interval="1m0s"

  # Quality of service level
  #
  # 0: at most once
  # 1: at least once
  # 2: exactly once
  #
  # Note: an increase of this value will decrease the performance.
  # For more information: https://www.hivemq.com/blog/mqtt-essentials-part-6-mqtt-quality-of-service-levels
  qos=0

  # Clean session
  #
  # Set the "clean session" flag in the connect message when this client
  # connects to an MQTT broker. By setting this flag you are indicating
  # that no messages saved by the broker for this client should be delivered.
  clean_session=true

  # Client ID
  #
  # Set the client id to be used by this client when connecting to the MQTT
  # broker. A client id must be no longer than 23 characters. When left blank,
  # a random id will be generated. This requires clean_session=true.
  client_id=""

  # CA certificate file (optional)
  #
  # Use this when setting up a secure connection (when server uses ssl://...)
  # but the certificate used by the server is not trusted by any CA certificate
  # on the server (e.g. when self generated).
  ca_cert=""

  # TLS certificate file (optional)
  tls_cert=""

  # TLS key file (optional)
  tls_key=""


  # AMQP / RabbitMQ.
  [application_server.integration.amqp]
  # Server URL.
  #
  # See for a specification of all the possible options:
  # https://www.rabbitmq.com/uri-spec.html
  url="amqp://guest:guest@localhost:5672"

  # Event routing key template.
  #
  # This is the event routing-key template used when publishing device
  # events.
  event_routing_key_template="application.{{ .ApplicationID }}.device.{{ .DevEUI }}.event.{{ .EventType }}"


  # AWS Simple Notification Service (SNS)
  [application_server.integration.aws_sns]
  # AWS region.
  #
  # Example: "eu-west-1".
  # See also: https://docs.aws.amazon.com/general/latest/gr/rande.html.
  aws_region=""

  # AWS Access Key ID.
  aws_access_key_id=""

  # AWS Secret Access Key.
  aws_secret_access_key=""

  # Topic ARN (SNS).
  topic_arn=""


  # Azure Service-Bus integration.
  [application_server.integration.azure_service_bus]
  # Connection string.
  #
  # The connection string can be found / created in the Azure console under
  # Settings -> Shared access policies. The policy must contain Manage & Send.
  connection_string=""

  # Publish mode.
  #
  # Select either "topic", or "queue".
  publish_mode=""

  # Publish name.
  #
  # The name of the topic or queue.
  publish_name=""


  # Google Cloud Pub/Sub integration.
  [application_server.integration.gcp_pub_sub]
  # Path to the IAM service-account credentials file.
  #
  # Note: this service-account must have the following Pub/Sub roles:
  #  * Pub/Sub Editor
  credentials_file=""

  # Google Cloud project id.
  project_id=""

  # Pub/Sub topic name.
  topic_name=""


  # Kafka integration.
  [application_server.integration.kafka]
  # Brokers, e.g.: localhost:9092.
  brokers=["localhost:9092"]

  # Topic for events.
  topic="chirpstack_as"

  # Template for keys included in Kafka messages. If empty, no key is included.
  # Kafka uses the key for distributing messages over partitions. You can use
  # this to ensure some subset of messages end up in the same partition, so
  # they can be consumed in-order. And Kafka can use the key for data retention
  # decisions.  A header "event" with the event type is included in each
  # message. There is no need to parse it from the key.
  event_key_template="{{ .ApplicationServer.Integration.Kafka.EventKeyTemplate }}"


  # PostgreSQL database integration.
  [application_server.integration.postgresql]
  # PostgreSQL dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable).
  dsn=""

  # This sets the max. number of open connections that are allowed in the
  # PostgreSQL connection pool (0 = unlimited).
  max_open_connections=0

  # Max idle connections.
  #
  # This sets the max. number of idle connections in the PostgreSQL connection
  # pool (0 = no idle connections are retained).
  max_idle_connections=2


  # Settings for the "internal api"
  #
  # This is the API used by ChirpStack Network Server to communicate with ChirpStack Application Server
  # and should not be exposed to the end-user.
  [application_server.api]
  # ip:port to bind the api server
  bind="0.0.0.0:8001"

  # ca certificate used by the api server (optional)
  ca_cert=""

  # tls certificate used by the api server (optional)
  tls_cert=""

  # tls key used by the api server (optional)
  tls_key=""

  # Public ip:port of the application-server API.
  #
  # This is used by ChirpStack Network Server to connect to ChirpStack Application Server. When running
  # ChirpStack Application Server on a different host than ChirpStack Network Server, make sure to set
  # this to the host:ip on which ChirpStack Network Server can reach ChirpStack Application Server.
  # The port must be equal to the port configured by the 'bind' flag
  # above.
  public_host="localhost:8001"


  # Settings for the "external api"
  #
  # This is the API and web-interface exposed to the end-user.
  [application_server.external_api]
  # ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api)
  bind="0.0.0.0:8080"

  # http server TLS certificate (optional)
  tls_cert=""

  # http server TLS key (optional)
  tls_key=""

  # JWT secret used for api authentication / authorization
  # You could generate this by executing 'openssl rand -base64 32' for example
  jwt_secret=""

  # Allow origin header (CORS).
  #
  # Set this to allows cross-domain communication from the browser (CORS).
  # Example value: https://example.com.
  # When left blank (default), CORS will not be used.
  cors_allow_origin=""


  # Settings for the remote multicast setup.
  [application_server.remote_multicast_setup]
  # Synchronization interval.
  sync_interval="1s"

  # Synchronization retries.
  sync_retries=3

  # Synchronization batch-size.
  sync_batch_size=100


  # Settings for the fragmentation-session setup.
  [application_server.fragmentation_session]
  # Synchronization interval.
  sync_interval="1s"

  # Synchronization retries.
  sync_retries=3

  # Synchronization batch-size.
  sync_batch_size=100



# Join-server configuration.
#
# ChirpStack Application Server implements a (subset) of the join-api specified by the
# LoRaWAN Backend Interfaces specification. This API is used by ChirpStack Network Server
# to handle join-requests.
[join_server]
# ip:port to bind the join-server api interface to
bind="0.0.0.0:8003"

# CA certificate (optional).
#
# When set, the server requires a client-certificate and will validate this
# certificate on incoming requests.
ca_cert=""

# TLS server-certificate (optional).
#
# Set this to enable TLS.
tls_cert=""

# TLS server-certificate key (optional).
#
# Set this to enable TLS.
tls_key=""


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
  as_kek_label=""

  # KEK set.
  #
  # Example (the [[join_server.kek.set]] can be repeated):
  # [[join_server.kek.set]]
  # # KEK label.
  # label="000000"

  # # Key Encryption Key.
  # kek="01020304050607080102030405060708"


# Metrics collection settings.
[metrics]
# Timezone
#
# The timezone is used for correctly aggregating the metrics (e.g. per hour,
# day or month).
# Example: "Europe/Amsterdam" or "Local" for the the system's local time zone.
timezone="Local"

  # Metrics stored in Redis.
  #
  # The following metrics are stored in Redis:
  # * gateway statistics
  [metrics.redis]
  # Aggregation intervals
  #
  # The intervals on which to aggregate. Available options are:
  # 'MINUTE', 'HOUR', 'DAY', 'MONTH'.
  aggregation_intervals=["MINUTE", "HOUR", "DAY", "MONTH"]

  # Aggregated statistics storage duration.
  minute_aggregation_ttl="2h0m0s"
  hour_aggregation_ttl="48h0m0s"
  day_aggregation_ttl="2160h0m0s"
  month_aggregation_ttl="17520h0m0s"


  # Metrics stored in Prometheus.
  #
  # These metrics expose information about the state of the ChirpStack Network Server
  # instance.
  [metrics.prometheus]
  # Enable Prometheus metrics endpoint.
  endpoint_enabled=false

  # The ip:port to bind the Prometheus metrics server to for serving the
  # metrics endpoint.
  bind=""

  # API timing histogram.
  #
  # By setting this to true, the API request timing histogram will be enabled.
  # See also: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
  api_timing_histogram=false


  # Monitoring settings.
  #
  # Note that this replaces the metrics.prometheus configuration. If a
  # metrics.prometheus if found in the configuration then it will fall back
  # to that and the monitoring section is ignored.
  [monitoring]

  # IP:port to bind the monitoring endpoint to.
  #
  # When left blank, the monitoring endpoint will be disabled.
  bind=""

  # Prometheus metrics endpoint.
  #
  # When set to true, Prometheus metrics will be served at '/metrics'.
  prometheus_endpoint=false

  # Prometheus API timing histogram.
  #
  # By setting this to true, the API request timing histogram will be enabled.
  # See also: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
  prometheus_api_timing_histogram=false

  # Health check endpoint.
  #
  # When set to true, the healthcheck endpoint will be served at '/health'.
  # When requesting, this endpoint will perform the following actions to
  # determine the health of this service:
  #   * Ping PostgreSQL database
  #   * Ping Redis database
  healthcheck_endpoint=false
{{< /highlight >}}

## Securing the application-server internal API

In order to protect the application-server internal API (`[application_server.internal_api]`) against
unauthorized access and to encrypt all communication, it is advised to use TLS
certificates. Once the `ca_cert`, `tls_cert` and `tls_key` are set, the
API will enforce client certificate validation on all incoming connections.
This means that when configuring a network-server instance in ChirpStack Application Server,
you must provide the CA and TLS client certificate in order to let the
network-server to connect to ChirpStack Application Server. See also
[network-server management]({{< ref "use/network-servers.md" >}}).

See [https://github.com/brocaar/chirpstack-certificates](https://github.com/brocaar/chirpstack-certificates)
for a set of script to generate such certificates.

## Securing the join-server API

In order to protect the join-server API (`[join_server]`) against
unauthorized access and to encrypt all communication, it is advised to use TLS
certificates. Once the `ca_cert`, `tls_cert` and `tls_key` are
set, the API will enforce client certificate validation on all incoming connections.
When the `ca_cert` is left blank, TLS will still be configured, but the server
will not require and validate the client-certificate.

Please note that you also need to configure ChirpStack Network Server so that it uses a
client certificate for its join-server API client. See
[ChirpStack Network Server configuration](https://www.chirpstack.io/network-server/install/config/).

## Securing the web-interface and external API

The web-interface and public api (`[application_server.external_api]`) can be
secured using a TLS certificate and key. Once the `tls_cert` and `tls_key`
are set (`[application_server.external_api]`), TLS will be activated.

### Self-signed certificate

A self-signed certificate can be generated with the following command:

{{<highlight bash>}}
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 90 -nodes
{{< /highlight >}}

### Let's Encrypt

For generating a certificate with [Let's Encrypt](https://letsencrypt.org/),
first follow the [getting started](https://letsencrypt.org/getting-started/)
instructions. When the `letsencrypt` cli tool has been installed, execute:

{{<highlight bash>}}
letsencrypt certonly --standalone -d DOMAINNAME.HERE 
{{< /highlight >}}
