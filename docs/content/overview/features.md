---
title: Features
menu:
    main:
        parent: overview
        weight: 2
---

# Features

LoRa App Server is an application-server, part of the LoRa Server project.
For features related to the network-server component, see the
[LoRa Server documentation](/loraserver/).

## Payload encryption / decryption

LoRa App Server handles the encryption and decryption of the application
payloads. It also holds the application-key of each device and handles the
join-accept in case of OTAA activation. This means that payloads will be
sent decrypted to the integrations, but also that before payloads are sent
to [LoRa Server](/loraserver/) meaning the network-server does not have access
to these payloads.

## Web-interface

LoRa App Server offers a web-interface (built on top of the provided
[RESTful]({{<ref "integrate/rest.md">}}) api). This web-interface can be used
to manage users, organizations, applications and devices.

## User authorization

Using LoRa App Server, it is possible to grant users global admin permissions,
make them admin of an organization or assign them view-only permissions within
an organization. This makes it possible to run LoRa App Server in a multi-tenant
environment where each organization or team has access to only their own applications
and devices.

## API

For intgration with external services, LoRa App server provides a [RESTFul]({{<ref "integrate/rest.md">}})
and [gRPC]({{<ref "integrate/grpc.md">}}) API which exposes the same
functionality as the web-interface. [Authentication and authorization]({{<ref "integrate/auth.md">}})
is implemented using JWT tokens.

## Payloads and device events

By default, LoRa App Server offers a MQTT integration for all configured
devices. The provided MQTT topics can be for receiving data from your devices,
sending downlink data or to get notified about events like joins, acks and
errors. See [Sending and receiving data]({{<ref "integrate/data.md">}}) for
more information.

Additional to the MQTT integration, it is possible to configure HTTP endpoints
for receiving device payloads and events. See
[Data integrations]({{<ref "integrate/integrations.md">}}) for more information.

**Note:** downlink payloads can also be scheduled through the API.

## Gateway discovery

For networks containing multiple gateways, LoRa App Server provides a feature
to test the gateway network coverage. By sending out periodical "pings" through
each gateway, LoRa App Server is able to discover how well these are received by
other gateways in the same network. The collected data is displayed as a map
in the web-interface.

This feature can be enabled and configured per [network-server]({{<ref "use/network-servers.md">}}).

## Live frame-logging

With LoRa App Server you are able to inspect all raw and encrypted LoRaWAN
frames per gateway or device. When opening the *Live LoRaWAN frame logs* tab on the
gateway or device detail page, you will see all frames passing in realtime.
This will also allow you to inspect the (encrypted) content of each LoRaWAN
frame. See [frame-logging]({{<ref "use/frame-logging.md">}}) for more information.

## Live event-logging

With LoRa App Server you are able to inspect all events from the web-interface,
without the need to use a MQTT client or build an integration. When opening
the *Live event logs* tab on the device detail pace, you will see all
uplink, ack, join and error events in realtime. See [event-logging]({{<ref "use/event-logging.md">}})
for more information.
