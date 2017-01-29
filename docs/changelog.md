# Changelog

## Development

**Features:**

* More descriptive error on missing `--http-tls-cert` and `--http-tls-key`
  configuration.

**Bugfixes:**

* `limit` and `offset` url parameters are now correctly documented in the
  API console for the `/api/node` endpoint.

## 0.2.0

**Features:**

* Adaptive data-rate support. See [loraserver/features](https://docs.loraserver.io/loraserver/features/)
  for more information about ADR. Note:

	* [LoRa Server](https://docs.loraserver.io/loraserver/) 0.13.0 or higher
	  is required
	* ADR is currently only implemented for the EU 863-870 ISM band
	* This is an experimental feature

* Besides RX information, TX information is exposed for received uplink
  payloads. See [MQTT topics](mqtt-topics.md) for more information.


## 0.1.4

* Update exposed data which is published to the
  `application/[AppEUI]/node/[DevEUI]/rx` topic. SNR, RSSI and GPS time (if
  available) are now included of all receiving gateways.
* Change `GW_MQTT_SERVER` environment variable to `MQTT_SERVER` (thanks @siscia).

## 0.1.3

* Make the relax frame-counter option visible in the ui (the static files
  were not generated and included in 0.1.2).

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
