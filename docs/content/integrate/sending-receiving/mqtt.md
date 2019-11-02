---
title: MQTT
menu:
    main:
        parent: sending-receiving
---

# MQTT integration

The MQTT integration publishes all the data it receices from the devices
as JSON over MQTT. To receive data from your device, you therefore 
need to subscribe to its MQTT topic. For debugging, you could use a 
(command-line) tool like `mosquitto_sub` which is part of the 
[Mosquitto](http://mosquitto.org/) MQTT broker.

Use `+` for a single-level wildcard, `#` for a multi-level wildcard.
Examples:

{{<highlight bash>}}
mosquitto_sub -t "application/123/#" -v          # display everything for the given application ID
mosquitto_sub -t "application/123/device/+/rx" -v  # display only the RX payloads for the given application ID
{{< /highlight >}}

**Notes:**

* MQTT topics are case-sensitive
* The `ApplicationID` can be retrieved using the API or from the web-interface,
  this is not the `AppEUI`!

## Events

The MQTT integration exposes all events as documented by [Event Types](../#event-types).

## Event MQTT topics

The following mapping to MQTT topics applies for the available events:

* Uplink: `application/[applicationID]/device/[devEUI]/rx`
* Status: `application/[applicationID]/device/[devEUI]/status`
* Ack: `application/[applicationID]/device/[devEUI]/ack`
* Error: `application/[applicationID]/device/[devEUI]/error`

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

## Scheduling downlink data

### application/[applicationID]/device/[devEUI]/tx

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

**Note:** the application ID and DevEUI of the device will be taken from the topic.

Example payload:

{{<highlight json>}}
{
    "confirmed": true,                        // whether the payload must be sent as confirmed data down or not
    "fPort": 10,                              // FPort to use (must be > 0)
    "data": "...."                            // base64 encoded data (plaintext, will be encrypted by ChirpStack Network Server)
    "object": {                               // decoded object (when application coded has been configured)
        "temperatureSensor": {"1": 25},       // when providing the 'object', you can omit 'data'
        "humiditySensor": {"1": 32}
    }
}

{{< /highlight >}}
