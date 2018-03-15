---
title: LoRa App Server
menu:
    main:
        parent: overview
        weight: 1
---

# LoRa App Server

LoRa App Server is an open-source LoRaWAN application-server, part of the
[LoRa Server](https://docs.loraserver.io/) project. It is responsible
for the device "inventory" part of a LoRaWAN infrastructure, handling of
join-request and the handling and encryption of application payloads.

It offers a [web-interface]({{<ref "use/login.md">}}) where users,
organizations, applications and devices can be managed. For integration with
external services, it offers a [RESTful]({{<ref "integrate/rest.md">}}) 
and [gRPC]({{<ref "integrate/grpc.md">}}) API.

By default application payloads and device events are [published over MQTT]({{<ref "integrate/data.md">}}),
but it is also possible through [additional integrations]({{<ref "integrate/integrations.md">}}) to publish payloads
and events to HTTP endpoints.

See also the complete list of [LoRa App Server features]({{<relref "features.md">}}).

## Screenshots

![applications](/lora-app-server/img/web_applications.png)
![nodes](/lora-app-server/img/web_nodes.png)
![node details](/lora-app-server/img/web_node_details.png)
![swagger api](/lora-app-server/img/swagger.png)
