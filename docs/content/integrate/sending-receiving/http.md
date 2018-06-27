---
title: HTTP
menu:
    main:
        parent: sending-receiving
---

# HTTP integration

When configured, the HTTP integration will make `POST` requests
to the configured endpoints on the following events:

* Received uplink data
* Status notifications
* Join notifications
* ACK notifications
* Error notifications

The HTTP integration follows exaclty the same JSON data structure as the
data structures documented in the [MQTT integration]({{< relref "mqtt.md" >}})
documentation.
