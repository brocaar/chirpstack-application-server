# Configuration

To list all configuration options, start `lora-app-server` with the `--help`
flag. This will display:

```
GLOBAL OPTIONS:
   --postgres-dsn value     postgresql dsn (e.g.: postgres://user:password@hostname/database?sslmode=disable) (default: "postgres://localhost/loraserver?sslmode=disable") [$POSTGRES_DSN]
   --db-automigrate         automatically apply database migrations [$DB_AUTOMIGRATE]
   --migrate-node-sessions  migrate some of the node-session data to the application-server storage (run this once when migrating from loraserver 0.11.x) [$MIGRATE_NODE_SESSIONS]
   --redis-url value        redis url (default: "redis://localhost:6379") [$REDIS_URL]
   --mqtt-server value      mqtt server (default: "tcp://localhost:1883") [$GW_MQTT_SERVER]
   --mqtt-username value    mqtt server username (optional) [$MQTT_USERNAME]
   --mqtt-password value    mqtt server password (optional) [$MQTT_PASSWORD]
   --ca-cert value          ca certificate used by the api server (optional) [$CA_CERT]
   --tls-cert value         tls certificate used by the api server (optional) [$TLS_CERT]
   --tls-key value          tls key used by the api server (optional) [$TLS_KEY]
   --bind value             ip:port to bind the api server (default: "0.0.0.0:8001") [$BIND]
   --http-bind value        ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api) (default: "0.0.0.0:8080") [$HTTP_BIND]
   --http-tls-cert value    http server TLS certificate [$HTTP_TLS_CERT]
   --http-tls-key value     http server TLS key [$HTTP_TLS_KEY]
   --jwt-secret value       JWT secret used for api authentication / authorization (disabled when left blank) [$JWT_SECRET]
   --ns-server value        hostname:port of the network-server api server (default: "127.0.0.1:8000") [$NS_SERVER]
   --ns-ca-cert value       ca certificate used by the network-server client (optional) [$NS_CA_CERT]
   --ns-tls-cert value      tls certificate used by the network-server client (optional) [$NS_TLS_CERT]
   --ns-tls-key value       tls key used by the network-server client (optional) [$NS_TLS_KEY]
   --help, -h               show help
   --version, -v            print the version
```

Both cli arguments and environment-variables can be used to pass configuration
options.

## Database migrations

It is possible to apply the database-migrations by hand
(see [migrations](https://github.com/brocaar/lora-app-server/tree/master/migrations))
or let LoRa App Server migrate to the latest state automatically, by using
the `--db-automigrate` flag. Make sure that you always make a backup when
upgrading Lora App Server and / or applying migrations.

## Web-interface and client API

The web-interface must be protected by a TLS certificate, as this allows to
run the gRPC and RESTful JSON api together on one port (`--http-tls-*` flags).
