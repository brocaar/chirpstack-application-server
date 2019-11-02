---
title: ThingsBoards.io
menu:
  main:
    parent: sending-receiving
---

# ThingsBoard integration

When configured, the ThingsBoard integration will send device attributes
and telemetry to the configured [ThingsBoard](https://thingsboard.io/) instance.

* [ThingsBoard guides](https://thingsboard.io/docs/guides/)

## Requirements

Before this integration is able to write data to ThingsBoard, the uplink
payloads must be decoded. The payload codec can be configured per
[Device Profile]({{<ref "use/device-profiles.md">}}). To validate that the uplink
payloads are decoded, you can use the [live device event-log]({{<ref "use/event-logging.md">}})
feature. Decoded payload data will be available under the `object` key in
the JSON object.

ThingsBoard will generate a _Access Token_ per device. This token must be
configured as a [device variable]({{<ref "use/devices.md">}}) in ChirpStack Application Server. 
The variable must be named **ThingsBoardAccessToken**.

## Attributes

For each event, ChirpStack Application Server will update the ThingsBoard device with the
following attributes:

* application_id
* application_name
* dev_eui
* device_name

In case any [device tags]({{<ref "use/devices.md">}}) are configured for the
device in ChirpStack Application Server, these will also be added to the attributes.

## Telemetry

### Uplink

Decoded uplink data is prefixed with the **data_** prefix. Make sure to
configure a coded in the [Device Profile]({{<ref "use/device-profiles.md">}}).

### Device-status

Device-status is prefixed with the **status_** prefix. The interval of the
device-status requests can be configured through the [Service Profile]({{<ref "use/service-profiles.md">}}).

### Location

Location data is prefixed with the **location_** prefix. Please note that this
is **only** available in case geolocation capable gateways are being used and
ChirpStack Network Server is configured with geolocation support.
