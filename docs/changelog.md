# Changelog

## 0.1.2

* Add relax frame-counter option (this requires LoRa Server >= 0.12.3)

## 0.1.1

* Better labels + help text for ui components
* Redirect to JWT token form on authorization error
* Add small delay between http server start and gprc-gateway init (to work
  around internal connection errors)
* Store all used DevNonce values instead of max. 10.

## 0.1.0

Initial release. LoRa App Server requires LoRa Server 0.12.0+.

### Migrating data from LoRa Server

The database migrations of LoRa App Server are compatible with the database
schema of LoRa Server 0.11.0. Running `lora-app-server` with `--db-automigrate`
will forward migrate the database to the new structure.

Because of the decoupling of the inventory part from LoRa Server, some
data needs to be migrated from Redis into PostgreSQL. This process can
be automated by starting `lora-app-server` with the
`--migrate-node-sessions` flag. This must be executed after running the
database migrations (`--db-automigrate` flag, you can use both flags
together).

When running `lora-app-server` with `--migrate-node-sessions`, it will loop
through all nodes in the database (PostgreSQL) and backfill the `DevAddr`,
`NwkSKey` and `AppSKey` from the node-session (Redis).

Make sure you have a backup of both databases before starting any migration :-)
