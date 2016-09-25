# LoRa App Server documentation

LoRa App Server is an open-source LoRaWAN application-server, compatible
with [LoRa Server](https://github.com/brocaar/loraserver). It is responsible
for the node "inventory" part of a LoRaWAN infrastructure, handling of received
application payloads and the downlink application payload queue. It comes
with a web-interface and API (RESTful JSON and gRPC) and supports authorization
by using JWT tokens (optional). Received payloads are published over MQTT
and payloads can be enqueued by using MQTT or the API.

## Screenshots

![nodes](img/web_nodes.png)
![node details](img/web_node_details.png)
![swagger api](img/swagger.png)

## Downloads

Pre-compiled binaries are available for:

* Linux (including ARM / Raspberry Pi)
* OS X
* Windows

See [https://github.com/brocaar/lora-app-server/releases](https://github.com/brocaar/lora-app-server/releases)
for downloads. Source-code can be found at
[https://github.com/brocaar/lora-app-server](https://github.com/brocaar/lora-app-server).

## License

LoRa App Server is distributed under the MIT license. See also
[LICENSE](https://github.com/brocaar/lora-app-server/blob/master/LICENSE).
