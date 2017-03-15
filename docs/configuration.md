# Configuration

To list all configuration options, start `lora-app-server` with the `--help`
flag. This will display:

```
GLOBAL OPTIONS:
   --postgres-dsn value        postgresql dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable) (default: "postgres://localhost/loraserver?sslmode=disable") [$POSTGRES_DSN]
   --db-automigrate            automatically apply database migrations [$DB_AUTOMIGRATE]
   --migrate-node-sessions     migrate some of the node-session data to the application-server storage (run this once when migrating from loraserver 0.11.x) [$MIGRATE_NODE_SESSIONS]
   --redis-url value           redis url (e.g. redis://user:password@hostname/0) (default: "redis://localhost:6379") [$REDIS_URL]
   --mqtt-server value         mqtt server (e.g. scheme://host:port where scheme is tcp, ssl or ws) (default: "tcp://localhost:1883") [$MQTT_SERVER]
   --mqtt-username value       mqtt server username (optional) [$MQTT_USERNAME]
   --mqtt-password value       mqtt server password (optional) [$MQTT_PASSWORD]
   --ca-cert value             ca certificate used by the api server (optional) [$CA_CERT]
   --tls-cert value            tls certificate used by the api server (optional) [$TLS_CERT]
   --tls-key value             tls key used by the api server (optional) [$TLS_KEY]
   --bind value                ip:port to bind the api server (default: "0.0.0.0:8001") [$BIND]
   --http-bind value           ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api) (default: "0.0.0.0:8080") [$HTTP_BIND]
   --http-tls-cert value       http server TLS certificate [$HTTP_TLS_CERT]
   --http-tls-key value        http server TLS key [$HTTP_TLS_KEY]
   --jwt-secret value          JWT secret used for api authentication / authorization [$JWT_SECRET]
   --ns-server value           hostname:port of the network-server api server (default: "127.0.0.1:8000") [$NS_SERVER]
   --ns-ca-cert value          ca certificate used by the network-server client (optional) [$NS_CA_CERT]
   --ns-tls-cert value         tls certificate used by the network-server client (optional) [$NS_TLS_CERT]
   --ns-tls-key value          tls key used by the network-server client (optional) [$NS_TLS_KEY]
   --pw-hash-iterations value  the number of iterations used to generate the password hash (default: 100000) [$PW_HASH_ITERATIONS]
   --log-level value           debug=5, info=4, warning=3, error=2, fatal=1, panic=0 (default: 4) [$LOG_LEVEL]
   --help, -h                  show help
   --version, -v               print the version
```

Both cli arguments and environment-variables can be used to pass configuration
options.

## PostgreSQL connection string

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

## Redis connection string

For more information about the Redis URL format, see:
[https://www.iana.org/assignments/uri-schemes/prov/redis](https://www.iana.org/assignments/uri-schemes/prov/redis).

## Database migrations

It is possible to apply the database-migrations by hand
(see [migrations](https://github.com/brocaar/lora-app-server/tree/master/migrations))
or let LoRa App Server migrate to the latest state automatically, by using
the `--db-automigrate` flag. Make sure that you always make a backup when
upgrading Lora App Server and / or applying migrations.

## Web-interface and client API

The web-interface must be protected by a TLS certificate, as this allows to
run the gRPC and RESTful JSON api together on one port (`--http-tls-*` flags).
