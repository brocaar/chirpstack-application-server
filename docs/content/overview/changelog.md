---
title: Changelog
menu:
    main:
        parent: overview
        weight: 4
toc: false
description: Lists the changes per ChirpStack Application Server release, including steps how to upgrade.
---

# Changelog

## v3.9.0

### Features

#### Monitoring

The monitoring configuration has been updated so that it is possible to
configure both a [Prometheus](https://prometheus.io/) endpoint at `/metrics`
and healthcheck endpoint at `/health`. This change is backwards compatible,
but to use the `/health` endpoint you must update your configuration.

#### Gateway tags and metadata

This makes it possible to store additional user-defined key/value configuration
per gateway. Metadata pushed by the [ChirpStack Gateway Bridge](https://www.chirpstack.io/gateway-bridge/overview/)
is now stored by the ChirpStack Application Server and visible under metadata.

#### Device-profile tags

This makes it possible to store additional user-defined key/value configuration
per device-profile.

#### myDevices Cayenne

The [myDevices Cayenne](https://developers.mydevices.com/cayenne/features/)
endpoint is now enabled for the myDevices integration.

### Improvements

* Implement searching devices by tags (e.g. `tagname:tagvalue`).
* Refactor Prometheus endpoint and add `/health` endpoint.

### Bugfixes

* Update gRPC dependency to fix 'DNS name does not exist' error. ([#426](https://github.com/brocaar/chirpstack-application-server/issues/426))

## v3.8.0

### Features

#### New error events

This release provide new error events for OTAA errors, frame-counter resets and
re-transactions / replay-attacks (requires ChirpStack Network Server v3.7+).
It also adds new error events when the gateway reports an error on downlink
scheduling (e.g. collision, invalid frequency, ...).

#### Downlink tx ack event

This publishes an acknowledgement when an item from the queue has been sent
to the gateway for transmission and was accepted by the gateway.

#### myDevices integration

This new integration makes it possible to forward payload events to myDevices
IoT in a Box. myDevices Cayenne integration will be made available soon.

#### Syslog logger

This adds a configuration option `log_to_syslog`. When enabled, log items will
be forwarded to syslog.

### Improvements

* Add API option to fetch device queue size only.
* Add device profile to device list in web-interface.
* Enable DNS round-robin load-balancing for gRPC client (to NS).
* Add uplink IDs used for gelocation to geolocation event. ([#413](https://github.com/brocaar/chirpstack-application-server/issues/413))
* Internal cleanup of unused `device_activation` table and code.
* Add option for 200 and 500 items per page in web-interface. ([#418](https://github.com/brocaar/chirpstack-application-server/pull/418))

### Bugfixes

* Fix LoRaWAN 1.1 OTAA bug (and update tests). ([#382](https://github.com/brocaar/chirpstack-application-server/issues/382))
* Fix assigning device profile to device (validate they share the same organization). ([#415](https://github.com/brocaar/chirpstack-application-server/issues/415))
* Fix assigning service profile to application (validate they share the same organization).

## v3.7.0

### Features

#### AMQP / RabbitMQ integration

This new integration publishes device events to an AMQP / [RabbitMQ](https://www.rabbitmq.com/)
message broker.

#### Device variables in JS codec

This makes it possible to store (for example) calibration values as
device variables and use this variable in the JavaScript codec.

### Improvements

#### API auth query cleanup

This cleanup optimizes the API authorization functions by using a SQL query
specific to each function, instead of a shared SQL query joining all tables.

#### Connection settings

* Implement max. reconnect interval setting for MQTT integration.
* Fix Redis max. idle configuration and add max. active connections configuration.
* Implement max open and idle connections for PostgreSQL integration.

### Bugfixes

* Check for null values in JSONTree web-interface component. ([#398](https://github.com/brocaar/chirpstack-application-server/issues/398))

## v3.6.1

### Bugfixes

* Fix setting `time` in `rxInfo` for `json_v3` (default) integration marshaler.
* Fix integrations break when codec function returns a NaN value. ([#390](https://github.com/brocaar/chirpstack-application-server/issues/390))

## v3.6.0

### Features

#### Protobuf based integrations

This feature adds a `marshaler` configuration option to the integration
configuration. The default is backwards compatible with the current JSON
format. New options are `protobuf` (Protobuf binary) and `json` (Protobuf based
JSON).

#### RPM packaging

This is the first release providing .rpm packages for CentOS and RedHat. ([#383](https://github.com/brocaar/chirpstack-application-server/pull/383)

### Improvements

#### gRPC / Protobuf cleanup

All definitions are now imported from `github.com/brocaar/chirpstack-api/go`.
When using the gRPC API, you must update your imports.

#### Azure Service-Bus integration

Implement re-connect in case of an Azure Service-Bus error. This to work
partly around [azure-service-bus-go/issues/149](https://github.com/Azure/azure-service-bus-go/issues/149).

#### Support multiple HTTP endpoints

Implement support for configuring multiple endpoints in the HTTP integration.

#### Internal improvements

Handlink of the received uplinks has been improved to reduce the number of
required database transactions.

### Bugfixes

* Payload codec is hidden on application create (it was already on edit).

## v3.5.1

### Improvements

* Deprecate use of dots (`.`) in environment variable names, use double underscore (`__`) instead. ([#369](https://github.com/brocaar/chirpstack-application-server/issues/369))

### Bugfixes

* Fixes init stop script which could cause the ChirpStack Application Server to not properly stop or restart. ([#379](https://github.com/brocaar/chirpstack-application-server/issues/379))

## v3.5.0

This release renames LoRa App Server to ChirpStack Application Server.
See the [Rename Announcement](https://www.chirpstack.io/r/rename-announcement) for more information.

### Improvements

* Add copy to clipboard for hex heys. ([#364](https://github.com/brocaar/chirpstack-applocation-server/issues/364))
* Persist selected number of table rows per page in localStorage. ([#351](https://github.com/brocaar/chirpstack-application-server/issues/351))

### Bugfixes

* Fix rendering empty gateway map.

## v3.4.0

### Features

#### IDs for correlation

This release implements per context unique IDs that are printed in the
logs and are returned as header in API responses. This makes it easier to
correlate log events.

#### Organization admin permissions

Next to the organization admin permission, this release makes it possible
to also (instead of full organization admin permissions), make an user
"device admin" or "gateway admin".

#### Migrate gateway stats

This release migrates the gateway stats from LoRa Server into LoRa App Server.
This also adds new configuration options to the configuration file, to configure
the timezone for aggregation and metrics retention.

#### Gateway map

In the gateway list page, it is now possible to show a map showing all the
gateways and their states.

### Improvements

* Add PostgreSQL max open / idle connections settings. ([#360](https://github.com/brocaar/lora-app-server/pull/360)
* Cleanup old freegeoip.net code for gateway location.

### Bugfixes

* Fix gateway last-seen in web-interface.

### Upgrading

This release requires LoRa Server v3.3.0 or later (which will forward the
gateway stats to the Application Server API). After upgrading LoRa App Server
and restarting it, it will run a migration to import existing stats from the
LoRa Server database. Therefore make sure LoRa Server is up and running
during the upgrade of LoRa App Server.

## v3.3.1

### Improvements

* Add RXInfo and TXInfo to join-notifications. ([#235](https://github.com/brocaar/chirpstack-application-server/pull/347))

### Bugfixes

* Propagate gRPC errors to that the correct HTTP status is returned. ([#349](https://github.com/brocaar/chirpstack-application-server/issues/349))
* Fix not found UI error when device is inactive. ([#359](https://github.com/brocaar/chirpstack-application-server/issues/359))

## v3.3.0

### Features

#### Multi-frame geolocation

The multi-frame geolocation (implemented by LoRa Server v3.2.0) fields have
been added to the Device Profile form in the web-interface.

#### Prometheus metrics

Prometheus metics have been implemented for the Join Server API interface.

### Improvements

* Update UI dependencies to their latest versions.
* Show popup on gateway create when no Service Profile exists. ([#345](https://github.com/brocaar/chirpstack-application-server/issues/345))
* Remove unused fields when listing users. ([#328](https://github.com/brocaar/chirpstack-application-server/issues/345))
* Update multicast address and session-key input field widgets.
* Mention in API description that the organization ID of an application can not be updated after create. ([#343](https://github.com/brocaar/chirpstack-application-server/issues/343))

### Bugfixes

* Trigger organization dropdown reload on organization change. ([#342](https://github.com/brocaar/chirpstack-application-server/issues/342))

## v3.2.0

### Features

#### Device queue

The LoRa App Server web-interface has now the option to:

* Enqueue payloads
* List pending queue items
* Flush the device queue

#### Prometheus metrics

gRPC API metrics can now be exposed using a [Prometheus](https://prometheus.io/) metrics endpoint.
In future releases, more metrics will be exposed using this endpoint.

### Improvements

* Multicast-group frame-counter cleanup (there was a column in both the LoRa App Server and LoRa Server database). ([#339](https://github.com/brocaar/chirpstack-application-server/issues/339))

### Bugfixes

* Fix missing multicast enqueue error handler (this caused the enqueue endpoint to return a 200 even in case of error).

## v3.1.0

### Features

#### Device variables and tags

Device variables and tags are user-defined key/value values that can be
assigned to devices. Tags can be used to add additional meta-data to devices
and variables can be used for configured integrations (e.g. ThingsBoard
requires a per-device Access Token).

#### PostgreSQL integration

When configured in the `lora-app-server.toml` configuration file, this
integration will write events into a PostgreSQL database. See the
[PostgreSQL Integration](https://www.chirpstack.io/application-server/integrate/sending-receiving/postgresql/)
documentation for more information.

#### ThingsBoard integration

When configured as Application integration, this integration will write
attributes and telemetry to the configured ThingsBoard instance. See the
[ThingsBoard Integration](https://www.chirpstack.io/application-server/integrate/sending-receiving/thingsboard/)
documentation for more information.

### Improvements

* Speedup login with default admin / admin credentials (for low CPU power devices). ([#320](https://github.com/brocaar/chirpstack-application-server/issues/320))
* Gateway ID, FPGA ID and AES-key widgets have been updated.
* Always display Class-C timeout field in Device Profile.

### Bugfixes

* Fix Gen Application Key is no longer a mandatory field. ([#322](https://github.com/brocaar/chirpstack-application-server/issues/322))
* Fix create organization user. ([#323](https://github.com/brocaar/chirpstack-application-server/issues/323))

### Upgrading

Before upgrading, you must enable the `hstore` extension for the LoRa App Server
PostgreSQL database. Example commands assuming the database is named `chirpstack_as`:

{{<highlight bash>}}
sudo -u postgres psql
{{< /highlight >}}

Within the PostgreSQL prompt, enter the following queries:

{{<highlight sql>}}
-- change to chirpstack_as database and enable extension
\c chirpstack_as
create extension hstore;

-- exit the prompt
\q
{{< /highlight >}}

## v3.0.0

### Features

#### Firmware Update Over The Air (FUOTA)

This release provides an implementation of the FUOTA specification.
Currently the deployment is limited to one device as we would like to gather
feedback on this implementation first. However, in this implementation the deployment
to groups of devices is taken into account.
**Note:** this feature is experimental and the API might change.

#### Updated rxInfo / txInfo (live frame-logs)

The LoRa Server <> LoRa Gateway Bridge messages have been updated to make the
downlink scheduling more flexible and generic. This change exposes these fields
in the live gateway and device frame-logs.

#### Payload codec

The payload codec settings have been moved to the Device Profile. Codec settings set
in the application configuration still remain functional, but new codec settings must be
configured in the Device Profile.

### Bugfixes

* Fix organization admin cannot add existing users. ([#315](https://github.com/brocaar/chirpstack-application-server/issues/315))

### Upgrading

Please upgrade LoRa Server first to v3 (see [Changelog](https://www.chirpstack.io/application-server/overview/changelog/)),
then upgrade LoRa App Server to v3. This LoRa App Server release is fully
backwards compatible.

## v2.6.1

### Improvements

* Make it possible in the API to move devices between applications.
* Add `DevAddr` to enqueue API call to LoRa Server so that LoRa Server can validate the session-keys are in sync.
* Add device details overview (UI).

### Bugfixes

* Fix `panic` when JOSN object is set but contains `null`. ([#314](https://github.com/brocaar/chirpstack-application-server/issues/314))

## v2.6.0

### Features

#### New integrations

Support has been added to forward events to an [Azure service-bus](https://azure.microsoft.com/en-us/services/service-bus/)
or to [AWS SNS](https://aws.amazon.com/sns/). See [Configuration](https://www.chirpstack.io/application-server/install/config/).

### Improvements

* Make JS codec maximum execution-time configurable and increase default value to 100ms.
* Add configuration option for CORS headers. ([#275](https://github.com/brocaar/chirpstack-application-server/pull/275))
* Internal code-cleanup with regards to passing configuration and objects.
* Internal migration from Dep to [Go modules](https://github.com/golang/go/wiki/Modules).
* Improve authentication validator SQL query for speed. ([#302](https://github.com/brocaar/chirpstack-application-server/pull/302))
* Add codec execution time for decoding. ([#307](https://github.com/brocaar/chirpstack-application-server/pull/307))
* UI: Make delete application confirmation more explicit. ([#306](https://github.com/brocaar/chirpstack-application-server/pull/306))

## v2.5.1

### Bugfixes

* Fix panic in InfluxDB handler on `null` values in object ([#295](https://github.com/brocaar/chirpstack-application-server/issues/295))

## v2.5.0

### Features

#### Support for retained messages

It is now possible to [configure](https://www.chirpstack.io/application-server/install/config/) the retained flag for the MQTT integration.
When set, the MQTT broker will retain the last message and send this
immediately when a MQTT client connects. ([#272](https://github.com/brocaar/chirpstack-application-server/pull/272))

#### Environment variable based configuration

Environment variable based [configuration](https://www.chirpstack.io/application-server/install/config/) has been re-implemented.

### Improvements

* Calls made by the HTTP integration are now made async.
* The alignment of the UI tabs has been improved.

### Bugfixes

* Fix potential deadlock on MQTT re-connect ([#103](https://github.com/brocaar/chirpstack-gateway-bridge/issues/103))
* Fix logrotate issue (init based systems) ([#282](https://github.com/brocaar/chirpstack-application-server/pull/282)

## v2.4.1

### Bugfixes

* Fix `createLeafletElement` implementation error (introduced by v2.4.0 leaflet upgrade).

## v2.4.0

### Improvements

#### TLS for web-interface and API optional

It is no longer required to configure a TLS certificate for securing the
LoRa App Server web-interface and API. This configuration is now optional
and unset by default.

#### InfluxDB uplink meta-data

The following values have been added:

* RSSI
* SNR
* Uplink frame-counter

#### EUI and key input fields (web-interface)

The device EUI and (session)key input fields have been improved for easier
input, supporting both MSB and LSB byte order. Also only the required fields
(based on LoRaWAN 1.0.x or 1.1.x) are displayed in the forms.

## v2.3.0

### Features

#### Google Cloud Platform integration

LoRa App Server is now able to publish application data to
[Cloud Pub/Sub](https://cloud.google.com/pubsub/) as an alternative to a MQTT
broker. Please refer to the [Configuration](https://www.chirpstack.io/application-server/install/config/)
for more information.

### Deactivate device API

An API endpoint has been added to de-activate (not remove) devices.

### Device battery status

LoRa App Server now publishes the device battery-level as a percentage instead
of a value between `0...255`. The `battery` field will be removed in the next
major release.

### Improvements

* The join-server `ca_cert` can be left blank to disable client-certificate
  validation when a TLS certificate is configured.

## v2.2.0

### Upgrade notes

This upgrade is backwards compatible with previous v2 releases, but when using
geolocation-support, you must also upgrade LoRa Server to v2.2.0+.

### Features

#### Geolocation

This release adds geolocation support.

* Configuration of fine-timestamp decryption keys (e.g. for the Kerlink iBTS).
* `.../location` MQTT topics on which device locations are published.
* Location notification endpoint for HTTP integration.
* Per device reference altitude (for more accurate geolocation).

#### Improvements

* Replace `garyburd/redigo/redis` with `gomodule/redigo/redis`.


#### Bugfixes

* Status notification endpoint was missing for HTTP integration.
* Fix `/api` endpoint redirecting to web-interface (this might require a clear cache).

## v2.1.0

### Upgrade notes

This upgrade is backwards compatible with previous v2 releases, but when using
multicast-support, you must also upgrade LoRa Server to v2.1.0+.

### Features

#### Multicast support

This adds experimental support for creating multicast-groups to which devices
can be assigned (potentially covered by multiple gateways).

#### LoRaWAN 1.0.3

This update adds LoRaWAN 1.0.3 in MAC version dropdown.

### Bugfixes

* Fix organization selector dropdown styling.

## v2.0.1

### Bugfixes

* Use `gofrs/uuid` UUID library as `satori/go.uuid` is not truly random. ([#253](https://github.com/brocaar/chirpstack-application-server/pull/253))
* Fix web-interface login form (sometimes a double login was required).

## v2.0.0

### Upgrade notes

Before upgrading to v2, first make sure you have the latest v1 installed and running
(including LoRa App Server). As always, it is recommended to make a backup
first :-)

### Features

#### LoRaWAN 1.1 support

This release adds support for LoRaWAN 1.1 devices (meaning that both LoRaWAN 1.0
and LoRaWAN 1.1 devices are supported). Please note that the LoRaWAN 1.0 *AppKey*
is now called *NwkKey* and LoRaWAN 1.1 adds a new key called *AppKey*.

#### (Encrypted) key signaling

The LoRa App Server join-server API supports using Key Encryption Keys (KEK)
for encrypting the session-keys on a (re)join-request, requested by LoRa Server.
It will also send the (encrypted) AppSKey in this response to LoRa Server.

When LoRa Server receives the first uplink from the device (in case of a rejoin-request,
this will be the first uplink using the new security context), it will send this
(encrypted) AppSKey together with the application payload to LoRa App Server.
This will also be the moment when LoRa App Server will sent the join notification!

#### New UI

The LoRa App Server web-interface has been re-designed with a focus on better
navigation. All main components are now accessible from a sidebar.

### Changes

#### Device-status
The device-status has been removed from the uplink payload and is sent over
a separate MQTT topic (or HTTP integration). This to make sure that the
the device-status is only published when an update is available.
See also [Sending and receiving data](https://www.chirpstack.io/application-server/integrate/sending-receiving/).

#### API changes

The API has been cleaned up to improve consistency and usability. This update
affects most of the endpoints! Most of these changes can be summarized by
the following example (where `device` is a separate object which now can be
re-used for create / get and update methods).

##### Old API

`POST /api/devices`

{{<highlight json>}}
{
  "name": "test-device",
  "devEUI": "0102030405060708",
  "applicationID": "123"
  ...
}
{{< /highlight >}}

##### New API

`POST /api/devices`

{{<highlight json>}}
{
  "device": {
    "name": "test-device",
    "devEUI": "0102030405060708",
    "applicationID": "123"
    ...
  }
}
{{< /highlight >}}

#### InfluxDB changes

The `device_uplink` measurement `spreading_factor`, `bandwidth`, `modulation`
and `bitrate` tags are now replaced by a single `dr` tag.

#### Uplink message payload

The uplink message payload (used for MQTT and HTTP integrations) has been
modified slightly:

* It now contains a `dr` field indicating the used uplink data-rate.
* Location related fields of each `rxInfo` element has been moved inside a `location` object.
* `MAC` has been renamed to `gatewayID` for each `rxInfo` element.
* The `adr` field has been moved out of `txInfo` and moved into the root object.

#### Downlink queue changes

The `reference` field has been removed to simplify the downlink queue handling.
When using the REST or gRPC API interface, the response to an enqueue action
contains the frame-counter mapped with the downlink queue item. This
frame-counter then can be used to map the acknowledgement in case of a confirmed
downlink payload.

## v1.0.2

### Bugfixes

* Lock device row on downlink enqueue to avoid duplicated frame-counter values ([#245](https://github.com/brocaar/chirpstack-application-server/issues/245))

## v1.0.1

### Improvements

* `tls_cert` and `tls_key` are set automatically (again) when installing from `.deb` file.

## v1.0.0

This marks the first stable release! 

### Upgrade notes

* First make sure you have v0.21.1 installed and running (together with LoRa Server v0.26.3).
* As some configuration defaults have been changed (in the MQTT topic `node`
  has been replaced by `device`), make sure the old defaults are in you config
  file. To re-generate a configuration file while keeping your modifications, run:
  {{<highlight bash>}}
  lora-app-server -c lora-app-server-old.toml configfile > lora-app-server.toml
  {{< /highlight >}}
* You are now ready to upgrade to v1.0.0!

See [Downloads](https://www.chirpstack.io/application-server/overview/downloads/)
for pre-compiled binaries or instructions how to setup the Debian / Ubuntu
repository for v1.x.

### Changes

* In the MQTT topic configuration defaults, `node` has been replaced by `device`.
* Code to remain backwards compatible with environment-variable based
  configuration has been removed.
* Code to create device- and service-profiles on upgrade from v0.14.0 has been removed.
* Code to migrate the device-queue on upgrade from v0.15.0 has been removed.
* Code to create gateway-profiles on upgrade from v0.20.0 has been removed.
* Old unused tables (kept for upgrade migration code) have been removed from db.

## 0.21.1

**Bugfixes:**

* Fix InfluxDB handler error for unexported fields (these are now skipped).
* Fix `data_data` in InfluxDB measurement names when using a JS based codec.
* Add missing `int64` and `uint64` value handling.

## 0.21.0

**Features:**

* LoRa App Server can now export decoded payload data directly to InfluxDB.
  See [Sending and receiving device data](https://www.chirpstack.io/application-server/integrate/sending-receiving/) for more information.

**Bugfixes:**

* In some cases the JavaScript editor (codec functions) would only render on click.

## 0.20.2

**Bugfixes:**

* JS codec now handles all possible int and float types returned by the encoder function.
* Fixed Gob decoding issues when decoding to `interface{}` by using JSON marshaler for logging events.

**Improvements:**
* JS codec downlink errors are now logged to `.../error` MQTT topic.

## 0.20.1

**Improvements:**

* Skip frame-counter check can now be set per device (so it can be used for OTAA devices).
* Publish codec decode errors to the `application/[applicationID]/node/[devEUI]/error` MQTT topic.

## 0.20.0

**Features:**

* (Gateway) channel-configuration has been refactored into gateway-profiles.
  * This requires [LoRa Server](https://www.chirpstack.io/application-server/) 0.26.0 or up.
  * This removes the channel-configuration related gateway API methods.
  * This adds gateway-profile API methods.

**Bugfixes:**

* Fix leaking Redis connections on pubsub subscriber ([#313](https://github.com/brocaar/chirpstack-network-server/issues/313).
* Fix discovery interval validation ([#226](https://github.com/brocaar/chirpstack-application-server/issues/226)).

**Upgrade notes:**

In order to automatically migrate the existing channel-configuration into the
new gateway-profiles, first upgrade LoRa Server and restart it. After upgrading
LoRa App Server and restarting it, all channel-configurations will be migrated
and associated to the gateways. As always, it is advised to first make a backup
of your (PostgreSQL) database.

## 0.19.0

**Features:**

* Global search on organizations, applications, devices and gateways.
* Display live device events (the same data as publised over MQTT).
  See also [Event logging](https://www.chirpstack.io/application-server/use/event-logging/).

**Improvements:**

* When creating an application, show a warning when no service-profile exists.
* When creating a gateway, show a warning when no network-server has been associated.
* When creating a device, show a warning when no device-profile exists.

**Bugfixes:**

* Fix organization selector (which would sometimes show an empty value on select).
* Fix user selector when assigning an user to an organization (which would sometimes show an empty value on select).

**Upgrade notes:**

Before upgrading, the PostgreSQL `pg_trgm` extension needs to be enabled.
Assuming the LoRa App Server database is configured as `chirpstack_as` this
extension could be enabled using the commands below.

Start the PostgreSQL prompt as the `postgres` user:

{{<highlight bash>}}
sudo -u postgres psql
{{< /highlight >}}

Within the PostgreSQL prompt, enter the following queries:

{{<highlight sql>}}
-- change to the LoRa App Server database
\c chirpstack_as

-- enable the extension
create extension pg_trgm;

-- exit the prompt
\q
{{< /highlight >}}

## 0.18.2

**Improvements:**

* Gateway discovery configuration has been moved to network-server configuration.
  * **Important:** when you have the gateway discover feature configured,
    you need to re-add this configuration under network-servers (web-interface).
* Expose the following MQTT options for the MQTT gateway backend:
  * Configurable MQTT topics (uplink, downlink, join, ack, error)
  * QoS (quality of service)
  * Client ID
  * Clean session on connect
* Expose LoRa Server version and configured region through the network-server
  API endpoint.
* Websocket client automatically re-connects on connection error ([#221](https://github.com/brocaar/chirpstack-application-server/pull/221))

**Bugfixes:**

* The Class-C enabled checkbox was displayed twice in the web-interface.
* Organization dropdown was not autocompleting correctly.

## 0.18.1

**Features:**

* Expose Class-B fields in device-profile web-interface form.
  * **Note:** Class-B support is implemented since [LoRa Server](https://www.chirpstack.io/network-server/) 0.25.0.

**Bugfixes:**

* Fix factory preset frequency field in device-profile form.

## 0.18.0

**Features:**

* LoRa App Server uses a new configuration file format.
  See [configuration](https://www.chirpstack.io/application-server/install/config/) for more information.
* Frame-logs for device are now streaming and can be downloaded as JSON file.
  * **Note:** the `/api/devices/{devEUI}/frames` (formerly `Device.GetFrameLogs`)
  endpoint has changed (and the gRPC method has been renamed to `Device.StreamFrameLogs`).
  * You need LoRa Server 0.24+ in order to use this feature.
* Added streaming frame-logs for gateways (which also can be downloaded as JSON file).
  * You need LoRa Server 0.24+ in order to use this feature.
* Support MQTT client certificate authentication ([#201](https://github.com/brocaar/chirpstack-application-server/pull/201)).

**Upgrade notes:**

When upgrading using the `.deb` package / using `apt` or `apt-get`, your
configuration will be automatically migrated for you. In any other case,
please see [configuration](https://www.chirpstack.io/application-server/install/config/).

## 0.17.1

**Bugfixes:**

* Fix missing `/` prefix in two UI links causing a redirect to the login page.
* Fix typo in TLS certificate loading causing error *failed to find certificate PEM data in certificate input* (thanks [@Francisco_Rivas](https://forum.chirpstack.io/u/Francisco_Rivas/summary))

## 0.17.0

**Features:**

* Device *last seen* timestamp is now stored and displayed in device list
* In the service-profile, it is now possible to set the
  * Device-status request frequency
  * Report battery level
  * Report margin

  When the interval is set to > 0 and reporting of this status is enabled,
  then this information is displayed in the device list and exposed over MQTT
  and the configured integrations.

* Extra logging has been added:
  * gRPC API calls (to the gRPC server and by the gRPC clients) are logged
    as `info`
  * Executed SQL queries are logged as `debug`
* A warning is displayed in the web-interface when creating a service-profile
  when no network-server is connected.
* A warning is displayed in the web-interface when creating a device-profile
  when the organization is not associated with a network-server.

**Internal changes:**

* The project moved to using [dep](https://github.com/golang/dep) as vendoring
  tool. In case you run into compiling issues, remove the `vendor` folder
  (which is not part of the repository anymore) and run `make requirements`.

* The front-end code has been updated to use React 16.2.0 and all dependencies
  have been updated.

**Bugfixes:**

* `--gw-ping-dr 0` is now handled correctly ([#204](https://github.com/brocaar/chirpstack-application-server/pull/204))


## 0.16.1


**Features:**

* Implement client certificate validation for incoming application-server API connections.
* Implement client certificate validation for incoming join-server API connections.
* Implement client certificate for API connections to LoRa Server.

This removes the following CLI options:

* `--ns-ca-cert`
* `--ns-tls-cert`
* `--ns-tls-key`

See for more information:

* [LoRa Server configuration](https://www.chirpstack.io/network-server/install/config/)
* [LoRa App Server configuration](https://www.chirpstack.io/application-server/install/config/)
* [LoRa App Server network-server management](https://www.chirpstack.io/application-server/use/network-servers/)
* [https://github.com/brocaar/chirpstack-network-server-certificates](https://github.com/brocaar/chirpstack-network-server-certificates)

**Improvements:**

* Optional note field (users) has been changed to textarea.
* Description field of the gateway has been changed to textarea.

**Bugfixes:**

* Fix device-profile permissions in UI for organization admins.
* Fix device-profile list per `applicationID` showing all device-profile names
  and IDs on the same network-server as the service-profile associated with the
  given application.

## 0.16.0

**Features:**

* LoRa App Server is now able to decode (uplink) and encode (downlink)
  payloads using the following per application configurable codecs:

  * None (only the raw `base64` encoded data will be exposed)
  * Cayenne LPP (data will be encoded / decoded using the
    [Cayenne LPP](https://mydevices.com/cayenne/docs/lora/) encoding)
  * Custom JavaScript codec functions (you can provide your own encoding /
    decoding functions in JavaScript)

See [Applications](https://www.chirpstack.io/application-server/use/applications/)
documentation for instructions how to configure this option.

## 0.15.0

**Changes:**

* Downlink device-queue

  * Downlink device-queue has been moved from the LoRa App Server database to
    the LoRa Server database.
  * LoRa App Server sends nACK when no confirmation has been received on
    confirmed downlink transmission. See [ACK notifications](https://www.chirpstack.io/application-server/integrate/data/).
  * LoRa App Server will not re-try transmitting a confirmed downlink anymore.
  * ACK and error notifications now contain the `fCnt` to which the notification is related.
  * The downlink-queue is now flushed on a (re)activation.

* Downlink device-queue API (`/api/devices/{devEUI}/queue`)
  * Removed `DELETE /api/devices/{devEUI}/queue/{id}` endpoint (as removing
    individual device-queue items will give `fCnt` gaps).
  * Added `DELETE /api/devices/{devEUI}/queue` to flush the whole device-queue.

* Class-C
  * Class-C timeout (see [device-profiles](https://www.chirpstack.io/application-server/use/device-profiles/))
    has been implemented for confirmed downlink transmissions. **Make sure to
    update this value for existing Class-C device-profiles to a sane value**.

**Bugfixes:**

* Fix organization pagination (thanks [@jcampanell-cablelabs](https://github.com/jcampanell-cablelabs))

**Improvements:**

* Use RFC1945 Authorization header format (thanks [@fifthaxe](https://github.com/fifthaxe))

### Upgrading

This release depends on LoRa Server 0.23.0. Upgrade LoRa Server first.
After upgrading LoRa App Server, it will migrate the remaining
device-queue items to the LoRa Server database.

## 0.14.2

**Bugfixes:**

* Fix unclosed response body (HTTP integrations).

## 0.14.1

**Bugfixes:**

* Remove `RxInfo` length validation as this slice is empty when
  *Add gateway meta-data* is disabled in the service-profile
  (thanks [@pni-jmattison](https://forum.chirpstack.io/u/pni-jmattison/summary)).
* Rename `/api/node/...` prefix of downlink queue into `/api/device/...`
  (thanks [@iegomez](https://github.com/iegomez)).
* Rename `DownlinkQueue...` gRPC methods and structs into `DeviceQueue...`.

## 0.14.0

**Note:** this release brings many changes! Make sure (as always) to make a
backup of your PostgreSQL and Redis database before upgrading.

**Changes:**

* Data-model refactor to implement service-profile, device-profile and
  routing-profile storage as defined in the
  [LoRaWAN backend interfaces](https://www.lora-alliance.org/lorawan-for-developers).

* Application users have been removed to avoid complexity in the API
  authorization. Users can still be assigned to organizations.

* LoRa App Server can now connect to multiple [LoRa Server](https://www.chirpstack.io/network-server/)
  instances.

* LoRa App Server exposes a Join-Server API (as defined in the LoRaWAN backend
  interfaces document), which LoRa Server uses as a default join-server.

* E-mail and note field added for users.

* Adaptive-datarate configuration has been moved to LoRa Server.

* OTAA RX configuration has been moved to LoRa Server.

**API changes:**

* New API endpoints:
  * `/api/device-profiles` (management of device-profiles)
  * `/api/service-profiles` (management of service-profiles)
  * `/api/network-servers` (management of network-servers)
  * `/api/devices` (management of devices, used to be `/api/nodes`, settings
    have been removed and device-profile field has been added)

* Updated API endpoints:
  * `/api/applications` (management of applications, most of the settings are now part of the device-profile)
  * `/api/gateways` (management of gateways, network-server field has been added)

* Removed API endpoints:
  * `/api/applications/{id}/users` (management of application users)
  * `/api/nodes` (management of nodes, has been refactored into `/api/devices`)

**Note:** these changes also apply to the related gRPC API endpoints.

### How to upgrade

**Note:** this release brings many changes! Make sure (as always) to make a
backup of your PostgreSQL and Redis database before upgrading.

**Note:** When LoRa App Server is running on a different server than LoRa Server,
make sure to set the `--as-public-server` / `AS_PUBLIC_SERVER`
(default `localhost:8001`).

This release depends on the latest LoRa Server release (0.22).
Start with updating LoRa Server first. See also the
[LoRa Server changelog](https://www.chirpstack.io/network-server/overview/changelog/).

LoRa App Server will perform the data-migration when the `--db-automigrate` / 
`DB_AUTOMIGRATE` config flag is set. It will:

* Create a network-server record + routing-profile on LoRa Server (so that
  LoRa Server knows how to connect back).
* For each organization, it will create a service-profile
* It will create device-profiles (either per device or per application when
  the "use application settings" is checked)

## 0.13.2

**Features:**

* The list of nodes can now be filtered on DevEUI or name prefix
  (thanks [@iegomez](https://github.com/iegomez)).

## 0.13.1

**Improvements:**

* Rename Gateway ping to Gateway discovery.
* Rename Frame logs to Raw frame logs and add note that these frames are encrypted.

## 0.13.0

**Features:**

* Gateway ping for testing the gateway coverage (by other gateways).
  When configured, LoRa App Server will send periodically pings through
  each gateway which has the ping functionality enabled.
  See also [features]{{<relref "features.md">}}.

Note: this release requires LoRa Server 0.21+ as the gateway ping feature
depends on the 'Proprietary' LoRaWAN message-type.

**Bugfixes:**

* Content-Type header was missing for HTTP integrations.

## 0.12.0

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

## 0.11.0

**Features:**

* Implement support for channel-configuration management. This makes it
  possible to assign channel-plans to gateways, which then can be used by
  [LoRa Gateway Config](https://docs.chirpstack.io/lora-gateway-config/).

**Note:** This feature is dependent on [LoRa Server](https://www.chirpstack.io/network-server/)
version 0.20.0+.

## 0.10.1

**Improvements:**

* `--mqtt-ca-cert` / `MQTT_CA_CERT` configuration flag was added to
  specify an optional CA certificate
  (thanks [@siscia](https://github.com/siscia)).

**Bugfixes:**

* MQTT client library update which fixes an issue where during a failed
  re-connect the protocol version would be downgraded
  ([paho.mqtt.golang#116](https://github.com/eclipse/paho.mqtt.golang/issues/116)).

## 0.10.0

**Features & changes:**

* Added frame logs tab to node to display all uplink and downlink
  frames of a given node. Note: This requires
  [LoRa Server](https://www.chirpstack.io/network-server/) 0.18.0.
* Updated organization, application and node navigation in UI.

## 0.9.1

**Bugfixes:**

* Fix ABP sesstings not editable by organization admin
  ([#85](https://github.com/brocaar/chirpstack-application-server/issues/85))

## 0.9.0

**Features & Changes:**

* Channel-lists have been removed. Extra channels (if supported by the ISM
  band) are now managed by LoRa Server configuration.

**Bugfixes:**

* On editing a gateway, disable the MAC input field (as this is the unique
  identifier of the gateway).
* A pagination regression has been fixed ([#82](https://github.com/brocaar/chirpstack-application-server/issues/82)).

**Note:** when upgrading to this version with `--db-automigrate` /
`DB_AUTOMIGRATE` set, channel-list data will be removed.

## 0.8.0

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

## 0.7.2

**Bugfixes:**

* Fix race-condition between fetching the gateway details and getting the
  current location if the gateway location is not yet set (UI).

## 0.7.1

**Features & changes:**

* Add 'set to current location' (create / update gateway in UI)

## 0.7.0

**Features & changes:**

* Gateway management and stats
    * Gateway management UI and API (accessible by global admin users) was added.
    * Gateway locations are now exposed in the [uplink MQTT topic]({{< ref "integrate/sending-receiving/mqtt.md" >}})
    * Requires [LoRa Server](/loraserver/) 0.16.0+.
* On MQTT and PostgreSQL connect error, LoRa App Server will retry instead
  of fail.

Many thanks again to [@jcampanell-cablelabs](https://github.com/jcampanell-cablelabs)
for collaborating on this feature.

## 0.6.0

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

## 0.5.0

**Features:**

* Per application (network) settings. Settings like Class-C, ADR, receive-window,
  data-rate etc. can now be managed on an application level. Optionally, they can
  be overridden per node.
* Applications and nodes are now paginated per 20 rows.

## 0.4.0

**Features & changes:**

* Class-C support. When adding / changing nodes, tick the Class-C checkbox
  to enable Class-C support for a given node.
* Application ID added to application overview (as it is needed to subscribe
  to MQTT topics).
* Nodes are now sorted by name.

**Note:** For Class-C functionality, upgrade LoRa Server to 0.14.0 or above.

## 0.3.0

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
  `DevEUI` (see [mqtt topics]({{< ref "integrate/sending-receiving/mqtt.md" >}}) for more info).
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

## 0.2.0

**Features:**

* Adaptive data-rate support. See [loraserver/features](https://www.chirpstack.io/network-server/features/)
  for more information about ADR. Note:

    * [LoRa Server](https://www.chirpstack.io/network-server/) 0.13.0 or higher
      is required
    * ADR is currently only implemented for the EU 863-870 ISM band
    * This is an experimental feature

* Besides RX information, TX information is exposed for received uplink
  payloads. See [MQTT topics]({{< ref "integrate/sending-receiving/mqtt.md" >}}) for more information.


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
