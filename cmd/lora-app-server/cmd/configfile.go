package cmd

import (
	"os"
	"text/template"

	"github.com/brocaar/lora-app-server/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// when updating this template, don't forget to update config.md!
const configTemplate = `[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}

# The number of times passwords must be hashed. A higher number is safer as
# an attack takes more time to perform.
password_hash_iterations={{ .General.PasswordHashIterations }}


# PostgreSQL settings.
#
# Please note that PostgreSQL 9.5+ is required.
[postgresql]
# PostgreSQL dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable).
#
# Besides using an URL (e.g. 'postgres://user:password@hostname/database?sslmode=disable')
# it is also possible to use the following format:
# 'user=loraserver dbname=loraserver sslmode=disable'.
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
# (see https://github.com/brocaar/lora-app-server/tree/master/migrations)
# or let LoRa App Server migrate to the latest state automatically, by using
# this setting. Make sure that you always make a backup when upgrading Lora
# App Server and / or applying migrations.
automigrate={{ .PostgreSQL.Automigrate }}


# Redis settings
#
# Please note that Redis 2.6.0+ is required.
[redis]
# Redis url (e.g. redis://user:password@hostname/0)
#
# For more information about the Redis URL format, see:
# https://www.iana.org/assignments/uri-schemes/prov/redis
url="{{ .Redis.URL }}"

# Max idle connections in the pool.
max_idle={{ .Redis.MaxIdle }}

# Idle timeout.
#
# Close connections after remaining idle for this duration. If the value
# is zero, then idle connections are not closed. You should set
# the timeout to a value less than the server's timeout.
idle_timeout="{{ .Redis.IdleTimeout }}"


# Application-server settings.
[application_server]
# Application-server identifier.
#
# Random UUID defining the id of the application-server installation (used by
# LoRa Server as routing-profile id).
# For now it is recommended to not change this id.
id="{{ .ApplicationServer.ID }}"


  # Integration configures the data integration.
  #
  # This is the data integration which is available for all applications,
  # besides the extra integrations that can be added on a per-application
  # basis.
  [application_server.integration]
  # Enabled integrations.
  #
  # Enabled integrations are enabled for all applications. Multiple
  # integrations can be configured.
  # Do not forget to configure the related configuration section below for
  # the enabled integrations. Integrations that can be enabled are:
  # * mqtt              - MQTT broker
  # * aws_sns           - AWS Simple Notification Service (SNS)
  # * azure_service_bus - Azure Service-Bus
  # * gcp_pub_sub       - Google Cloud Pub/Sub
  enabled=[{{ if .ApplicationServer.Integration.Enabled|len }}"{{ end }}{{ range $index, $elm := .ApplicationServer.Integration.Enabled }}{{ if $index }}", "{{ end }}{{ $elm }}{{ end }}{{ if .ApplicationServer.Integration.Enabled|len }}"{{ end }}]


  # MQTT integration backend.
  [application_server.integration.mqtt]
  # MQTT topic templates for the different MQTT topics.
  #
  # The meaning of these topics are documented at:
  # https://www.loraserver.io/lora-app-server/integrate/data/
  #
  # The following substitutions can be used:
  # * "{{ "{{ .ApplicationID }}" }}" for the application id.
  # * "{{ "{{ .DevEUI }}" }}" for the DevEUI of the device.
  #
  # Note: the downlink_topic_template must contain both the application id and
  # DevEUI substitution!
  uplink_topic_template="{{ .ApplicationServer.Integration.MQTT.UplinkTopicTemplate }}"
  downlink_topic_template="{{ .ApplicationServer.Integration.MQTT.DownlinkTopicTemplate }}"
  join_topic_template="{{ .ApplicationServer.Integration.MQTT.JoinTopicTemplate }}"
  ack_topic_template="{{ .ApplicationServer.Integration.MQTT.AckTopicTemplate }}"
  error_topic_template="{{ .ApplicationServer.Integration.MQTT.ErrorTopicTemplate }}"
  status_topic_template="{{ .ApplicationServer.Integration.MQTT.StatusTopicTemplate }}"
  location_topic_template="{{ .ApplicationServer.Integration.MQTT.LocationTopicTemplate }}"

  # Retained messages configuration.
  #
  # The MQTT broker will store the last publised message, when retained message is set
  # to true. When a client subscribes to a topic with retained message set to true, it will
  # always receive the last published message.
  uplink_retained_message={{ .ApplicationServer.Integration.MQTT.UplinkRetainedMessage }}
  join_retained_message={{ .ApplicationServer.Integration.MQTT.JoinRetainedMessage }}
  ack_retained_message={{ .ApplicationServer.Integration.MQTT.AckRetainedMessage }}
  error_retained_message={{ .ApplicationServer.Integration.MQTT.ErrorRetainedMessage }}
  status_retained_message={{ .ApplicationServer.Integration.MQTT.StatusRetainedMessage }}
  location_retained_message={{ .ApplicationServer.Integration.MQTT.LocationRetainedMessage }}

  # MQTT server (e.g. scheme://host:port where scheme is tcp, ssl or ws)
  server="{{ .ApplicationServer.Integration.MQTT.Server }}"

  # Connect with the given username (optional)
  username="{{ .ApplicationServer.Integration.MQTT.Username }}"

  # Connect with the given password (optional)
  password="{{ .ApplicationServer.Integration.MQTT.Password }}"

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


  # Settings for the "internal api"
  #
  # This is the API used by LoRa Server to communicate with LoRa App Server
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
  # This is used by LoRa Server to connect to LoRa App Server. When running
  # LoRa App Server on a different host than LoRa Server, make sure to set
  # this to the host:ip on which LoRa Server can reach LoRa App Server.
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

  # when set, existing users can't be re-assigned (to avoid exposure of all users to an organization admin)"
  disable_assign_existing_users={{ .ApplicationServer.ExternalAPI.DisableAssignExistingUsers }}

{{ if ne .ApplicationServer.Branding.Header  "" }}
  # Branding configuration.
  [application_server.branding]
  # Header
  header="{{ .ApplicationServer.Branding.Header }}"

  # Footer
  footer="{{ .ApplicationServer.Branding.Footer }}"

  # Registration.
  registration="{{ .ApplicationServer.Branding.Registration }}"

{{ end }}

# Join-server configuration.
#
# LoRa App Server implements a (subset) of the join-api specified by the
# LoRaWAN Backend Interfaces specification. This API is used by LoRa Server
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
# The KEK meganism is used to encrypt the session-keys sent from the
# join-server to the network-server. 
#
# The LoRa App Server join-server will use the NetID of the requesting
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
