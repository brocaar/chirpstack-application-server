---
title: MQTT
menu:
    main:
        parent: sending-receiving
---

# MQTT integration

The MQTT integration publishes all the data it receives from the devices
as JSON over MQTT. To receive data from your device, you therefore 
need to subscribe to its MQTT topic. For debugging, you could use a 
(command-line) tool like `mosquitto_sub` which is part of the 
[Mosquitto](http://mosquitto.org/) MQTT broker.

## Quickstart

Use `+` for a single-level wildcard, `#` for a multi-level wildcard.
Examples:

{{<highlight bash>}}
mosquitto_sub -t "application/123/#" -v                  # display everything for the given application ID
mosquitto_sub -t "application/123/device/+/event/up" -v  # display only the uplink payloads for the given application ID
{{< /highlight >}}

**Notes:**

* MQTT topics are case-sensitive
* The `ApplicationID` can be retrieved using the API or from the web-interface,
  this is not the `AppEUI`!

## Events

The MQTT integration exposes all events as documented by [Event Types](../#event-types).
The default event topic is: `application/[ApplicationID]/device/[DevEUI]/event/[EventType]`

**Note:** Before v3.11.0, the default event topic was: `application[ApplicationID]/device/[DevEUI]/[EventType]`.
In case these are configured in the ChirpStack Application Server configuration,
then these will override the default configuration.

## Scheduling a downlink

The default topic for scheduling downlink payloads is: `application/[ApplicationID]/device/[DevEUI]/command/down`.

**Note:** Before v3.11.0, the default event topic was: `application[ApplicationID]/device/[DevEUI]/tx`.
In case these are configured in the ChirpStack Application Server configuration,
then these will override the default configuration.

**Note:** The ApplicationID and DevEUI of the device will be taken from the topic.

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
