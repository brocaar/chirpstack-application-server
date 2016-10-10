# LoRa App Server

[![Build Status](https://travis-ci.org/brocaar/lora-app-server.svg?branch=master)](https://travis-ci.org/brocaar/lora-app-server)
[![GoDoc](https://godoc.org/github.com/brocaar/lora-app-server?status.svg)](https://godoc.org/github.com/brocaar/lora-app-server)
[![Gitter chat](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/loraserver/lora-app-server)

LoRa App Server is an open-source LoRaWAN application-server, compatible
with [LoRa Server](https://github.com/brocaar/loraserver). It is responsible
for the node "inventory" part of a LoRaWAN infrastructure, handling of received
application payloads and the downlink application payload queue. It comes
with a web-interface and API (RESTful JSON and gRPC) and supports authorization
by using JWT tokens (optional). Received payloads are published over MQTT
and payloads can be enqueued by using MQTT or the API.

## Screenshots

![nodes](docs/img/web_nodes.png)
![node details](docs/img/web_node_details.png)
![swagger api](docs/img/swagger.png)


## Documentation

Please refer to the [documentation](https://docs.loraserver.io/lora-app-server/) for the
[getting started](https://docs.loraserver.io/lora-app-server/getting-started/)
documentation and implemented [features](https://docs.loraserver.io/lora-app-server/features/).

## Downloads

Pre-compiled binaries are available for:

* Linux (including ARM / Raspberry Pi)
* OS X
* Windows

See [https://github.com/brocaar/lora-app-server/releases](https://github.com/brocaar/lora-app-server/releases)
for downloads. Source-code can be found at
[https://github.com/brocaar/lora-app-server](https://github.com/brocaar/lora-app-server).

## Building from source

Building from souce requires a working [Go](https://golang.org/) environment
with vendoring support. Alternatively, you could use the provided
[docker-compose](https://docs.docker.com/compose/) environment for compiling
the source.

## License

LoRa App Server is distributed under the MIT license. See also
[LICENSE](https://github.com/brocaar/lora-app-server/blob/master/LICENSE).
