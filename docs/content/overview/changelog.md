---
title: Changelog
menu:
    main:
        parent: overview
        weight: 4
---

## Changelog

### 0.13.0

**Features:**

* Gateway ping for testing the gateway coverage (by other gateways).
  When configured, LoRa App Server will send periodically pings through
  each gateway which has the ping functionality enabled.
  See also [features]{{<relref "features.md">}}.

**Bugfixes:**

* Content-Type header was missing for HTTP integrations.

### 0.12.0

**Features:**

* HTTP data integration. This makes it possible to setup per application
  http integrations (LoRa App Server posting to configurable HTTP endpoints).
  Note that LoRa Server will always send the data to the MQTT broker.

**Improvements:**

* Better pagination in case there are many pages (thanks [@iegomez](https://github.com/iegomez)).
* Various code has been cleaned up.

**Bugfixes:**

* Fixed duplicated resultset-items when requesting all applications within
  an organization.

### 0.11.0

**Features:**

* Implement support for channel-configuration management. This makes it
  possible to assign channel-plans to gateways, which then can be used by
  [LoRa Gateway Config](https://docs.loraserver.io/lora-gateway-config/).

**Note:** This feature is dependent on [LoRa Server](https://docs.loraserver.io/loraserver/)
version 0.20.0+.

### 0.10.1

**Improvements:**

* `--mqtt-ca-cert` / `MQTT_CA_CERT` configuration flag was added to
  specify an optional CA certificate
  (thanks [@siscia](https://github.com/siscia)).

**Bugfixes:**

* MQTT client library update which fixes an issue where during a failed
  re-connect the protocol version would be downgraded
  ([paho.mqtt.golang#116](https://github.com/eclipse/paho.mqtt.golang/issues/116)).

### 0.10.0

**Features & changes:**

* Added frame logs tab to node to display all uplink and downlink
  frames of a given node. Note: This requires
  [LoRa Server](https://docs.loraserver.io/loraserver/) 0.18.0.
* Updated organization, application and node navigation in UI.

### 0.9.1

**Bugfixes:**

* Fix ABP sesstings not editable by organization admin
  ([#85](https://github.com/brocaar/lora-app-server/issues/85))

### 0.9.0

**Features & Changes:**

* Channel-lists have been removed. Extra channels (if supported by the ISM
  band) are now managed by LoRa Server configuration.

**Bugfixes:**

* On editing a gateway, disable the MAC input field (as this is the unique
  identifier of the gateway).
* A pagination regression has been fixed ([#82](https://github.com/brocaar/lora-app-server/issues/82)).

**Note:** when upgrading to this version with `--db-automigrate` /
`DB_AUTOMIGRATE` set, channel-list data will be removed.

### 0.8.0

**Features & changes:**

* Organization management so that applications and gateways can be managed per
  organization.
* Node-session migrate function (to migrate from LoRa Server 0.11.*) has
  been cleaned up.

**Bugfixes:**

* Openstreetmap tiles can now be fetched using https:// (in some cases Safari
  was not loading any tiles because of mixed content).
* When the browser does not allow using the location service, freegeoip.net is
  used as a fallback.
* Map zoom in/out on scrolling has been disabled (to avoid accidental zoom on
  scrolling the page).

Many thanks again to [@jcampanell-cablelabs](https://github.com/jcampanell-cablelabs)
for collaborating on this feature.

### 0.7.2

**Bugfixes:**

* Fix race-condition between fetching the gateway details and getting the
  current location if the gateway location is not yet set (UI).

### 0.7.1

**Features & changes:**

* Add 'set to current location' (create / update gateway in UI)

### 0.7.0

**Features & changes:**

* Gateway management and stats
    * Gateway management UI and API (accessible by global admin users) was added.
    * Gateway locations are now exposed in the [uplink MQTT topic]({{< ref "integrate/data.md" >}})
    * Requires [LoRa Server](/loraserver/) 0.16.0+.
* On MQTT and PostgreSQL connect error, LoRa App Server will retry instead
  of fail.

Many thanks again to [@jcampanell-cablelabs](https://github.com/jcampanell-cablelabs)
for collaborating on this feature.

### 0.6.0

**This release contains changes that are not backwards compatible!**

**Features & changes:**

* User management with support to assign users to applications.
* API authentication / authorization (JWT token) format has been updated
  (it contains the username instead of all the permissions of the user).
  See [API documentation]({{< ref "integrate/api.md" >}}) for more information.
* `--jwt-secret` / `JWT_SECRET` is now mandatory.
* An initial user with *admin* / *admin* credentials will be created (change
  this password immediately!).
* Node-session API has been removed, use the `/api/nodes/{devEUI}/activation`
  endpoint for getting (and setting) node activations.
* Updated web-interface to support user management.
* New API endpoints for creating users and assigning users to applications.

Many thanks to [@jcampanell-cablelabs](https://github.com/jcampanell-cablelabs)
for collaborating on this feature.

### 0.5.0

**Features:**

* Per application (network) settings. Settings like Class-C, ADR, receive-window,
  data-rate etc. can now be managed on an application level. Optionally, they can
  be overridden per node.
* Applications and nodes are now paginated per 20 rows.

### 0.4.0

**Features & changes:**

* Class-C support. When adding / changing nodes, tick the Class-C checkbox
  to enable Class-C support for a given node.
* Application ID added to application overview (as it is needed to subscribe
  to MQTT topics).
* Nodes are now sorted by name.

**Note:** For Class-C functionality, upgrade LoRa Server to 0.14.0 or above.

### 0.3.0

**This release contains changes that are not backwards compatible!**

**Features & changes:**

* Nodes can now be grouped per application (e.g. called `temperature-sensor`).
  For backwards compatibility, the `AppEUI` is used as application name when
  upgrading.
* Nodes can now be given a name (e.g. `garden-sensor`), which must be unique 
  within an application. For backwards compatibility the `DevEUI` is used as
  name for the nodes when upgrading.
* Application ID, and the name of the application and node are included in the
  MQTT payloads.
* JWT token validation is now based on the ID of the application instead of the
  `AppEUI` (which should not be used for grouping nodes by purpose).
* The gRPC and REST apis have been updated to reflect the above application and
  node name changes.
* The MQTT topics (and payloads) are now based on the application ID and node
  `DevEUI` (see [mqtt topics]({{< ref "integrate/data.md" >}}) for more info).
* An endpoint for activating nodes and for fetching the activation status has
  been added (before this was done by using the node-session endpoint).
* The node activation mode can now be set (OTAA or ABP). Incoming join-requests
  for ABP nodes will be rejected.
* The node-session API is now accessible at `/api/node/{devEUI}/session`.

Many thanks to [@jcampanell-cablelabs](https://github.com/jcampanell-cablelabs)
and [@VirTERM](https://twitter.com/VirTERM) for their input on the API changes.

**Bugfixes:**

* `limit` and `offset` url parameters are now correctly documented in the
  API console.
* More descriptive error on missing `--http-tls-cert` and `--http-tls-key`
  configuration.

### 0.2.0

**Features:**

* Adaptive data-rate support. See [loraserver/features](https://docs.loraserver.io/loraserver/features/)
  for more information about ADR. Note:

    * [LoRa Server](https://docs.loraserver.io/loraserver/) 0.13.0 or higher
      is required
    * ADR is currently only implemented for the EU 863-870 ISM band
    * This is an experimental feature

* Besides RX information, TX information is exposed for received uplink
  payloads. See [MQTT topics]({{< ref "integrate/data.md" >}}) for more information.


### 0.1.4

* Update exposed data which is published to the
  `application/[AppEUI]/node/[DevEUI]/rx` topic. SNR, RSSI and GPS time (if
  available) are now included of all receiving gateways.
* Change `GW_MQTT_SERVER` environment variable to `MQTT_SERVER` (thanks @siscia).

### 0.1.3

* Make the relax frame-counter option visible in the ui (the static files
  were not generated and included in 0.1.2).

### 0.1.2

* Add relax frame-counter option (this requires LoRa Server >= 0.12.3)

### 0.1.1

* Better labels + help text for ui components
* Redirect to JWT token form on authorization error
* Add small delay between http server start and gprc-gateway init (to work
  around internal connection errors)
* Store all used DevNonce values instead of max. 10.

### 0.1.0

Initial release. LoRa App Server requires LoRa Server 0.12.0+.

#### Migrating data from LoRa Server

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
