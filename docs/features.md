# Features

## API

LoRa App Server provides both a [gRPC](http://www.grpc.io) and RESTful API for
easy integration with your own applications. Both interfaces can be secured
using [JWT](http://jwt.io/) to limit users to certain resources. See
the [API](api.md) documentation for more details.

## Web interface

On top of the provided API, LoRa App Server provides a web-interface for the
management of nodes and node settings. In case [JWT](https://jwt.io/) is
configured, the user is requested to enter his / her JWT token.

## Uplink data

Uplink data is published to a MQTT broker so that it is easy to subscribe
to received data. Received data will be decrypted by LoRa App Server before
being published. See also [MQTT topics](mqtt-topics.md) for more information.

## Downlink data

LoRa App Server keeps an internal persistent queue of payloads to send to 
[LoRa Server](https://docs.loraserver.io/loraserver/) when a downlink receive
window occurs. Downlink payloads can be added to the queue by publishing them
to the MQTT topic of the corresponding node. Just before sending the payload
to [LoRa Server](https://docs.loraserver.io/loraserver/), the payload will be
encrypted. See also [MQTT topics](mqtt-topics.md) for more information.

## Event notification

Besides uplink and downlink payload handling, LoRa App Server will publish also
events over MQTT. An event can be a node joining, a payload acknowledged by
a node or an error (e.g. a downlink payload that exceeded the maximum payload
size). See also [MQTT topics](mqtt-topics.md) for more information.
