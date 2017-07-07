---
title: LoRa App Server
menu:
    main:
        parent: overview
        weight: 1
---

## LoRa App Server

LoRa App Server is an open-source LoRaWAN application-server, compatible
with [LoRa Server](https://github.com/brocaar/loraserver). It is responsible
for the node "inventory" part of a LoRaWAN infrastructure, handling of received
application payloads and the downlink application payload queue. It comes
with a web-interface and API (RESTful JSON and gRPC) and supports authorization
by using JWT tokens (optional). Received payloads are published over MQTT
and payloads can be enqueued by using MQTT or the API.

### Screenshots

![applications](/lora-app-server/img/web_applications.png)
![nodes](/lora-app-server/img/web_nodes.png)
![node details](/lora-app-server/img/web_node_details.png)
![swagger api](/lora-app-server/img/swagger.png)
