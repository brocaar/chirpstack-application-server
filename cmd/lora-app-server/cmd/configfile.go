package cmd

import (
	"html/template"
	"os"

	"github.com/gusseleet/lora-app-server/internal/config"
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


# Application-server settings.
[application_server]
# Application-server identifier.
#
# Random UUID defining the id of the application-server installation (used by
# LoRa Server as routing-profile id).
# For now it is recommended to not change this id.
id="{{ .ApplicationServer.ID }}"


  # MQTT integration configuration used for publishing (data) events
  # and scheduling downlink application payloads.
  # Next to this integration which is always available, the user is able to
  # configure additional per-application integrations.
  [application_server.integration.mqtt]
  # MQTT server (e.g. scheme://host:port where scheme is tcp, ssl or ws)
  server="{{ .ApplicationServer.Integration.MQTT.Server }}"

  # Connect with the given username (optional)
  username="{{ .ApplicationServer.Integration.MQTT.Username }}"

  # Connect with the given password (optional)
  password="{{ .ApplicationServer.Integration.MQTT.Password }}"

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

  # http server TLS certificate
  tls_cert="{{ .ApplicationServer.ExternalAPI.TLSCert }}"

  # http server TLS key
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
  # Gateway discovery configuration.
  #
  # When enabled, each gateway will periodically broadcast a discovery "ping"
  # which other gateways in the network are able to receive and which is
  # presented in the web-interface as a map.
  [application_server.gateway_discovery]
  # Enable the gateway discovery feature.
  enabled={{ .ApplicationServer.GatewayDiscovery.Enabled }}

  # the interval used for each gateway to send a ping
  interval="{{ .ApplicationServer.GatewayDiscovery.Interval }}"

  # the frequency used for transmitting the gateway ping (in Hz)
  frequency={{ .ApplicationServer.GatewayDiscovery.Frequency }}

  # the data-rate to use for transmitting the gateway ping
  dr={{ .ApplicationServer.GatewayDiscovery.DR }}


# Join-server configuration.
#
# LoRa App Server implements a (subset) of the join-api specified by the
# LoRaWAN Backend Interfaces specification. This API is used by LoRa Server
# to handle join-requests.
[join_server]
# ip:port to bind the join-server api interface to
bind="{{ .JoinServer.Bind }}"

# ca certificate used by the join-server api server
ca_cert="{{ .JoinServer.CACert }}"

# tls certificate used by the join-server api server (optional)
tls_cert="{{ .JoinServer.TLSCert }}"

# tls key used by the join-server api server (optional)
tls_key="{{ .JoinServer.TLSKey }}"


# Network-server configuration.
#
# This configuration is only used to migrate from older LoRa App Server.
[network_server]
server="{{ .NetworkServer.Server }}"
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
