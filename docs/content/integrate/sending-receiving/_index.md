---
title: Sending / receiving data
menu:
    main:
        identifier: sending-receiving
        parent: integrate
        weight: 1
description: Sending to and receiving data from devices.
listPages: false
---

# Sending and receiving device data

## APIs

The easiest way to send data to a device is using the [gRPC]({{<ref "/integrate/grpc.md">}})
or [RESTful JSON]({{<ref "/integrate/rest.md">}}) API. After adding a payload to the queue,
the downlink frame-counter is directly returned which will be used on an acknowledgement
of a confirmed-downlink.

## Global integrations

Global integrations are configured throught the [lora-app-server.toml]({{<ref "install/config.md">}})
configuration file and are globally enabled. This means that for every application configured,
the application will be published using this / these integration(s).

The following integrations are available:

* [MQTT]({{<relref "mqtt.md">}})


## Application integrations

Additional to the global enabled integrations, it is also possible to setup
integrations per [application]({{<ref "use/applications.md">}}).

THe following integrations are available:

* [HTTP]({{<relref "http.md">}})
* [InfluxDB]({{<relref "influxdb.md">}})
