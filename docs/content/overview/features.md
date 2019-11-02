---
title: Features
menu:
    main:
        parent: overview
        weight: 2
description: Features provided by the ChirpStack Application Server.
---

# Features

ChirpStack Application Server is an open-source LoRaWAN<sup>&reg;</sup> Application
Server, part of the [ChirpStack](https://chirpstack.io) LoRaWAN Network Server stack.
For features related to the Network Server component, please refer to the
[ChirpStack Network Server documentation](/network-server/).

## Payload encryption / decryption

ChirpStack Application Server handles the encryption and decryption of the application
payloads. It also holds the application-key of each device and handles the
join-accept in case of OTAA activation. This means that payloads will be
sent decrypted to the integrations, but also that before payloads are sent
to [ChirpStack Network Server](/network-server/) meaning the Network Server does not
have access to these payloads.

## Web-interface

ChirpStack Application Server offers a web-interface (built on top of the provided
[RESTful]({{<ref "integrate/rest.md">}}) api). This web-interface can be used
to manage users, organizations, applications and devices.

## User authorization

Using ChirpStack Application Server, it is possible to grant users global admin permissions,
make them admin of an organization or assign them view-only permissions within
an organization. This makes it possible to run ChirpStack Application Server in a multi-tenant
environment where each organization or team has access to only their own applications
and devices.

## API

For integration with external services, ChirpStack Application Server provides a [RESTFul]({{<ref "integrate/rest.md">}})
and [gRPC]({{<ref "integrate/grpc.md">}}) API which exposes the same
functionality as the web-interface. [Authentication and authorization]({{<ref "integrate/auth.md">}})
is implemented using JWT tokens.

## Payloads and device events

ChirpStack Application Server provides different ways of sending and receiving device
payloads (e.g. MQTT, HTTP, InfluxDB, ...).
Please refer to [Sending and receiving](/application-server/integrate/sending-receiving/)
for all available integrations.

**Note:** downlink payloads can also be scheduled through the API.

## Gateway discovery

For networks containing multiple gateways, ChirpStack Application Server provides a feature
to test the gateway network coverage. By sending out periodical "pings" through
each gateway, ChirpStack Application Server is able to discover how well these are received by
other gateways in the same network. The collected data is displayed as a map
in the web-interface.

This feature can be enabled and configured per [Network Server]({{<ref "use/network-servers.md">}}).

## Live frame-logging

With ChirpStack Application Server you are able to inspect all raw and encrypted LoRaWAN
frames per gateway or device. When opening the *LoRaWAN frames* tab on the
gateway or device detail page, you will see all frames passing in realtime.
This will also allow you to inspect the (encrypted) content of each LoRaWAN
frame. See [Frame Logging]({{<ref "use/frame-logging.md">}}) for more information.

## Live event-logging

With ChirpStack Application Server you are able to inspect all events from the web-interface,
without the need to use a MQTT client or build an integration. When opening
the *Live event logs* tab on the device detail pace, you will see all
uplink, ack, join and error events in realtime. See [Event Logging]({{<ref "use/event-logging.md">}})
for more information.
