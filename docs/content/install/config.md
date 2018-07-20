---
title: Configuration
menu:
    main:
        parent: install
        weight: 4
---

# Configuration

The `lora-app-server` binary has the following command-line flags:

```text
LoRa App Server is an open-source application-server, part of the LoRa Server project
	> documentation & support: https://www.loraserver.io/lora-app-server
	> source & copyright information: https://github.com/brocaar/lora-app-server

Usage:
  lora-app-server [flags]
  lora-app-server [command]

Available Commands:
  configfile  Print the LoRa Application Server configuration file
  help        Help about any command
  version     Print the LoRa App Server version

Flags:
  -c, --config string   path to configuration file (optional)
  -h, --help            help for lora-app-server
      --log-level int   debug=5, info=4, error=2, fatal=1, panic=0 (default 4)

Use "lora-app-server [command] --help" for more information about a command.
```

## Configuration file

By default `lora-app-server` will look in the following order for a
configuration file at the following paths when `--config` is not set:

* `lora-app-server.toml` (current working directory)
* `$HOME/.config/lora-app-server/lora-app-server.toml`
* `/etc/lora-app-server/lora-app-server.toml`

To load configuration from a different location, use the `--config` flag.

To generate a new configuration file `lora-app-server.toml`, execute the following command:

```bash
lora-app-server configfile > lora-app-server.toml
```

Note that this configuration file will be pre-filled with the current configuration
(either loaded from the paths mentioned above, or by using the `--config` flag).
This makes it possible when new fields get added to upgrade your configuration file
while preserving your old configuration. Example:

```bash
lora-app-server configfile --config lora-app-server-old.toml > lora-app-server-new.toml
```

Example configuration file:

```toml
[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level=4

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
dsn="postgres://localhost/loraserver_as?sslmode=disable"

# Automatically apply database migrations.
#
# It is possible to apply the database-migrations by hand
# (see https://github.com/brocaar/lora-app-server/tree/master/migrations)
# or let LoRa App Server migrate to the latest state automatically, by using
# this setting. Make sure that you always make a backup when upgrading Lora
# App Server and / or applying migrations.
automigrate=true


# Redis settings
#
# Please note that Redis 2.6.0+ is required.
[redis]
# Redis url (e.g. redis://user:password@hostname/0)
#
# For more information about the Redis URL format, see:
# https://www.iana.org/assignments/uri-schemes/prov/redis
url="redis://localhost:6379"


# Application-server settings.
[application_server]
# Application-server identifier.
#
# Random UUID defining the id of the application-server installation (used by
# LoRa Server as routing-profile id).
# For now it is recommended to not change this id.
id="6d5db27e-4ce2-4b2b-b5d7-91f069397978"


  # MQTT integration configuration used for publishing (data) events
  # and scheduling downlink application payloads.
  # Next to this integration which is always available, the user is able to
  # configure additional per-application integrations.
  [application_server.integration.mqtt]
  # MQTT topic templates for the different MQTT topics.
  #
  # The meaning of these topics are documented at:
  # https://docs.loraserver.io/lora-app-server/integrate/data/
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

  # MQTT server (e.g. scheme://host:port where scheme is tcp, ssl or ws)
  server="tcp://localhost:1883"

  # Connect with the given username (optional)
  username=""

  # Connect with the given password (optional)
  password=""

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


  # Settings for the "internal api"
  #
  # This is the API used by LoRa Server to communicate with LoRa App Server
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
  # This is used by LoRa Server to connect to LoRa App Server. When running
  # LoRa App Server on a different host than LoRa Server, make sure to set
  # this to the host:ip on which LoRa Server can reach LoRa App Server.
  # The port must be equal to the port configured by the 'bind' flag
  # above.
  public_host="localhost:8001"


  # Settings for the "external api"
  #
  # This is the API and web-interface exposed to the end-user.
  [application_server.external_api]
  # ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api)
  bind="0.0.0.0:8080"

  # http server TLS certificate
  tls_cert=""

  # http server TLS key
  tls_key=""

  # JWT secret used for api authentication / authorization
  # You could generate this by executing 'openssl rand -base64 32' for example
  jwt_secret=""

  # when set, existing users can't be re-assigned (to avoid exposure of all users to an organization admin)"
  disable_assign_existing_users=false



# Join-server configuration.
#
# LoRa App Server implements a (subset) of the join-api specified by the
# LoRaWAN Backend Interfaces specification. This API is used by LoRa Server
# to handle join-requests.
[join_server]
# ip:port to bind the join-server api interface to
bind="0.0.0.0:8003"

# ca certificate used by the join-server api server
ca_cert=""

# tls certificate used by the join-server api server (optional)
tls_cert=""

# tls key used by the join-server api server (optional)
tls_key=""


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
  as_kek_label=""

  # KEK set.
  #
  # Example (the [[join_server.kek.set]] can be repeated):
  # [[join_server.kek.set]]
  # # KEK label.
  # label="000000"

  # # Key Encryption Key.
  # kek="01020304050607080102030405060708"
```

## Securing the application-server internal API

In order to protect the application-server internal API (`[application_server.internal_api]`) against
unauthorized access and to encrypt all communication, it is advised to use TLS
certificates. Once the `ca_cert`, `tls_cert` and `tls_key` are set, the
API will enforce client certificate validation on all incoming connections.
This means that when configuring a network-server instance in LoRa App Server,
you must provide the CA and TLS client certificate in order to let the
network-server to connect to LoRa App Server. See also
[network-server management]({{< ref "use/network-servers.md" >}}).

See [https://github.com/brocaar/loraserver-certificates](https://github.com/brocaar/loraserver-certificates)
for a set of script to generate such certificates.

## Securing the join-server API

In order to protect the join-server API (`[join_server]`) against
unauthorized access and to encrypt all communication, it is advised to use TLS
certificates. Once the `ca_cert`, `tls_cert` and `tls_key` are
set, the API will enforce client certificate validation on all incoming connections.

Please note that you also need to configure LoRa Server so that it uses a
client certificate for its join-server API client. See
[LoRa Server configuration](https://docs.loraserver.io/loraserver/install/config/).

## Web-interface and public API

The web-interface and public api (`[application_server.public_api]`) must be
secured by a TLS certificate and key, as this allows to run the gRPC and RESTful
JSON api together on one port.

### Self-signed certificate

A self-signed certificate can be generated with the following command:

```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 90 -nodes
```

### Let's Encrypt

For generating a certificate with [Let's Encrypt](https://letsencrypt.org/),
first follow the [getting started](https://letsencrypt.org/getting-started/)
instructions. When the `letsencrypt` cli tool has been installed, execute:

```bash
letsencrypt certonly --standalone -d DOMAINNAME.HERE 
```