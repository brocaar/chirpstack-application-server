# Features

## API

LoRa App Server provides both a [gRPC](http://www.grpc.io) and RESTful API for
easy integration with your own applications. Both interfaces are secured
using [JWT](http://jwt.io/) to limit users to certain resources. See
the [API](api.md) documentation for more details.

## Web interface

On top of the provided API, LoRa App Server provides a web-interface for the
management of users, applications and nodes.

## Users

Users can be granted (admin) access to certain applications.

## Uplink data

Uplink data is published to a MQTT broker so that it is easy to subscribe
to received data. Received data will be decrypted by LoRa App Server before
being published. See also [MQTT topics](mqtt-topics.md) for more information.

## Downlink data

LoRa App Server keeps an internal queue of payloads to be emitted to the nodes.
Items can be scheduled using the [API](api.md) or
[MQTT topics](mqtt-topics.md). Both confirmed as unconfirmed payloads are
supported. In case of a confirmed payload, an ACK will be sent over MQTT
(see [MQTT topics](mqtt-topics.md)). Payloads will be encrypted by LoRa App
Server before they are sent to [LoRa Server](https://docs.loraserver.io/loraserver/).

### Class-A

Received downlink queue items will be added to the queue, awaiting the next
downlink receive window (the node initiate this by sending an uplink). In
case there are multiple queue items, it will indicate the node that there
is more data, so that the node can initiate a new receive window.

### Class-C

In case of Class-C, no internal queue will be kept. Received downlink queue
items are directly emitted to the node. In case of a confirmed payload, it will
be kept in the queue (with pending state) until an acknowledgement has been
received from the node.

## Event notification

Besides uplink and downlink payload handling, LoRa App Server will publish also
events over MQTT. An event can be a node joining, a payload acknowledged by
a node or an error (e.g. a downlink payload that exceeded the maximum payload
size). See also [MQTT topics](mqtt-topics.md) for more information.
