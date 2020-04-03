---
title: ChirpStack Application Server
menu:
    main:
        parent: overview
        weight: 1
listPages: false
---

# ChirpStack Application Server

ChirpStack Application Server is an open-source LoRaWAN<sup>&reg;</sup>
Application Server, part of the [ChirpStack](https://www.chirpstack.io/) open-source
LoRaWAN Network Server stack. It is responsible for the device "inventory"
part of a LoRaWAN infrastructure, handling of join-request and the handling
and encryption of application payloads.

It offers a [web-interface]({{<ref "use/login.md">}}) where users,
organizations, applications and devices can be managed. For integration with
external services, it offers a [RESTful]({{<ref "integrate/rest.md">}}) 
and [gRPC]({{<ref "integrate/grpc.md">}}) API.

Device data can be [sent and / or received](/application-server/integrate/sending-receiving/) over
MQTT, HTTP and be written directly into InfluxDB.

See also the complete list of [ChirpStack Application Server features]({{<relref "features.md">}}).

## Screenshots

![applications](/application-server/img/web_applications.png)
![nodes](/application-server/img/web_nodes.png)
![node details](/application-server/img/web_node_details.png)
![swagger api](/application-server/img/swagger.png)
