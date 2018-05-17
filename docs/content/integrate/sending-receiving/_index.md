---
title: Sending / receiving data
menu:
    main:
        identifier: sending-receiving
        parent: integrate
        weight: 1
---

# Sending and receiving device data

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
