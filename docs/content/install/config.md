---
title: Configuration
menu:
    main:
        parent: install
        weight: 4
---

## Configuration

To list all configuration options, start `lora-app-server` with the `--help`
flag. This will display:

```text
GLOBAL OPTIONS:
   --postgres-dsn value             postgresql dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable) (default: "postgres://localhost/loraserver?sslmode=disable") [$POSTGRES_DSN]
   --db-automigrate                 automatically apply database migrations [$DB_AUTOMIGRATE]
   --redis-url value                redis url (e.g. redis://user:password@hostname/0) (default: "redis://localhost:6379") [$REDIS_URL]
   --mqtt-server value              mqtt server (e.g. scheme://host:port where scheme is tcp, ssl or ws) (default: "tcp://localhost:1883") [$MQTT_SERVER]
   --mqtt-username value            mqtt server username (optional) [$MQTT_USERNAME]
   --mqtt-password value            mqtt server password (optional) [$MQTT_PASSWORD]
   --mqtt-ca-cert value             mqtt CA certificate file used by the gateway backend (optional) [$MQTT_CA_CERT]
   --mqtt-cert value                mqtt certificate file used by the gateway backend (optional) [$MQTT_CERT]
   --mqtt-cert-key value            mqtt key file of certificate used by the gateway backend (optional) [$MQTT_CERT_KEY]
   --as-public-server value         ip:port of the application-server api (used by LoRa Server to connect back to LoRa App Server) (default: "localhost:8001") [$AS_PUBLIC_SERVER]
   --as-public-id value             random uuid defining the id of the application-server installation (used by LoRa Server as routing-profile id) (default: "6d5db27e-4ce2-4b2b-b5d7-91f069397978") [$AS_PUBLIC_ID]
   --bind value                     ip:port to bind the api server (default: "0.0.0.0:8001") [$BIND]
   --ca-cert value                  ca certificate used by the api server (optional) [$CA_CERT]
   --tls-cert value                 tls certificate used by the api server (optional) [$TLS_CERT]
   --tls-key value                  tls key used by the api server (optional) [$TLS_KEY]
   --http-bind value                ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api) (default: "0.0.0.0:8080") [$HTTP_BIND]
   --http-tls-cert value            http server TLS certificate [$HTTP_TLS_CERT]
   --http-tls-key value             http server TLS key [$HTTP_TLS_KEY]
   --jwt-secret value               JWT secret used for api authentication / authorization [$JWT_SECRET]
   --pw-hash-iterations value       the number of iterations used to generate the password hash (default: 100000) [$PW_HASH_ITERATIONS]
   --log-level value                debug=5, info=4, warning=3, error=2, fatal=1, panic=0 (default: 4) [$LOG_LEVEL]
   --disable-assign-existing-users  when set, existing users can't be re-assigned (to avoid exposure of all users to an organization admin) [$DISABLE_ASSIGN_EXISTING_USERS]
   --gw-ping                        enable sending gateway pings [$GW_PING]
   --gw-ping-interval value         the interval used for each gateway to send a ping (default: 24h0m0s) [$GW_PING_INTERVAL]
   --gw-ping-frequency value        the frequency used for transmitting the gateway ping (in Hz) (default: 0) [$GW_PING_FREQUENCY]
   --gw-ping-dr value               the data-rate to use for transmitting the gateway ping (default: 0) [$GW_PING_DR]
   --js-bind value                  ip:port to bind the join-server api interface to (default: "0.0.0.0:8003") [$JS_BIND]
   --js-ca-cert value               ca certificate used by the join-server api server (optional) [$JS_CA_CERT]
   --js-tls-cert value              tls certificate used by the join-server api server (optional) [$JS_TLS_CERT]
   --js-tls-key value               tls key used by the join-server api server (optional) [$JS_TLS_KEY]
   --help, -h                       show help
   --version, -v                    print the version
```

Both cli arguments and environment-variables can be used to pass configuration
options.

### PostgreSQL connection string

Besides using an URL (e.g. `postgres://user:password@hostname/database?sslmode=disable`)
it is also possible to use the following format:
`user=loraserver dbname=loraserver sslmode=disable`.

The following connection parameters are supported:

* dbname - The name of the database to connect to
* user - The user to sign in as
* password - The user's password
* host - The host to connect to. Values that start with / are for unix domain sockets. (default is localhost)
* port - The port to bind to. (default is 5432)
* sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
* fallback_application_name - An application_name to fall back to if one isn't provided.
* connect_timeout - Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
* sslcert - Cert file location. The file must contain PEM encoded data.
* sslkey - Key file location. The file must contain PEM encoded data.
* sslrootcert - The location of the root certificate file. The file must contain PEM encoded data.

Valid values for sslmode are:

* disable - No SSL
* require - Always SSL (skip verification)
* verify-ca - Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
* verify-full - Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)

### Redis connection string

For more information about the Redis URL format, see:
[https://www.iana.org/assignments/uri-schemes/prov/redis](https://www.iana.org/assignments/uri-schemes/prov/redis).

### Database migrations

It is possible to apply the database-migrations by hand
(see [migrations](https://github.com/brocaar/lora-app-server/tree/master/migrations))
or let LoRa App Server migrate to the latest state automatically, by using
the `--db-automigrate` flag. Make sure that you always make a backup when
upgrading Lora App Server and / or applying migrations.

### Securing the application-server API

In order to protect the application-server API (listening on `--bind`) against
unauthorized access and to encrypt all communication, it is advised to use TLS
certificates. Once the `--ca-cert`, `--tls-cert` and `--tls-key` are set, the
API will enforce client certificate validation on all incoming connections.
This means that when configuring a network-server instance in LoRa App Server,
you must provide the CA and TLS client certificate in order to let the
network-server to connect to LoRa App Server. See also
[network-server management]({{< ref "use/network-servers.md" >}}).

See [https://github.com/brocaar/loraserver-certificates](https://github.com/brocaar/loraserver-certificates)
for a set of script to generate such certificates.

### Securing the join-server API

In order to protect the join-server API (listening on `--js-bind`) against
unauthorized access and to encrypt all communication, it is advised to use TLS
certificates. Once the `--js-ca-cert`, `--js-tls-cert` and `--js-tls-key` are
set, the API will enforce client certificate validation on all incoming connections.

Please note that you also need to configure LoRa Server so that it uses a
client certificate for its join-server API client. See
[LoRa Server configuration](https://docs.loraserver.io/loraserver/install/config/).

### Web-interface and client API

The web-interface must be secured by a TLS certificate, as this allows to
run the gRPC and RESTful JSON api together on one port (`--http-tls-*` flags).

#### Self-signed certificate

A self-signed certificate can be generated with the following command:

```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 90 -nodes
```

#### Let's Encrypt

For generating a certificate with [Let's Encrypt](https://letsencrypt.org/),
first follow the [getting started](https://letsencrypt.org/getting-started/)
instructions. When the `letsencrypt` cli tool has been installed, execute:

```bash
letsencrypt certonly --standalone -d DOMAINNAME.HERE 
```

### Gateway discovery

By configuring the `--gw-ping` / `GW_PING` settings LoRa App Server will
emit periodical gateway pings to test the coverage of each gateway. Make sure
that the `--gw-ping-frequency` / `GW_PING_FREQUENCY` setting is set to a
frequency that is part of the channel-plan of the other receiving gateways.

### Application Server public host

When running LoRa App Server on a different host than LoRa Server, make sure
to set the `--as-public-server` to the correct `hostname:port` on which LoRa
Server can reach LoRa App Server. The port must be equal to the port as
configured by the `--bind` / `BIND` configuration.
