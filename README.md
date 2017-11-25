# LoRa App Server

[![Build Status](https://travis-ci.org/brocaar/lora-app-server.svg?branch=master)](https://travis-ci.org/brocaar/lora-app-server)

LoRa App Server is an open-source LoRaWAN application-server, part of the
[LoRa Server](https://docs.loraserver.io/) project. It is responsible
for the node "inventory" part of a LoRaWAN infrastructure, handling of received
application payloads and the downlink application payload queue. It comes
with a web-interface and API (RESTful JSON and gRPC) and supports authorization
by using JWT tokens (optional). Received payloads are published over MQTT
and payloads can be enqueued by using MQTT or the API.

## Architecture

![architecture](https://docs.loraserver.io/img/architecture.png)

### Component links

* [LoRa Gateway Bridge](https://docs.loraserver.io/lora-gateway-bridge)
* [LoRa Gateway Config](https://docs.loraserver/lora-gateway-config)
* [LoRa Server](https://docs.loraserver.io/loraserver/)
* [LoRa App Server](https://docs.loraserver.io/lora-app-server/)

## Links

* [Downloads](https://docs.loraserver.io/lora-app-server/overview/downloads/)
* [Docker image](https://hub.docker.com/r/loraserver/lora-app-server/)
* [Documentation & screenshots](https://docs.loraserver.io/lora-app-server/) and [Getting started](https://docs.loraserver.io/lora-app-server/getting-started/)
* [Building from source](https://docs.loraserver.io/lora-app-server/community/source/)
* [Contributing](https://docs.loraserver.io/lora-app-server/community/contribute/)
* Support
  * [Support forum](https://forum.loraserver.io)
  * [Bug or feature requests](https://github.com/brocaar/lora-app-server/issues)

## Sponsors

[![CableLabs](https://www.loraserver.io/img/sponsors/cablelabs.png)](https://www.cablelabs.com/)
[![SIDNFonds](https://www.loraserver.io/img/sponsors/sidn_fonds.png)](https://www.sidnfonds.nl/)
[![acklio](https://www.loraserver.io/img/sponsors/acklio.png)](http://www.ackl.io/)

## License

LoRa App Server is distributed under the MIT license. See also
[LICENSE](https://github.com/brocaar/lora-app-server/blob/master/LICENSE).
