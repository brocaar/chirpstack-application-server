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

## Enqueueing downlink payloads

The easiest way to send data to a device is using the [gRPC]({{<ref "/integrate/grpc.md">}})
or [RESTful JSON]({{<ref "/integrate/rest.md">}}) API. After adding a payload to the queue,
the downlink frame-counter is directly returned which will be used on an acknowledgement
of a confirmed-downlink.

## Data integrations

### Global integrations

Global integrations are configured throught the [lora-app-server.toml]({{<ref "install/config.md">}})
configuration file and are globally enabled. This means that for every application configured,
the application will be published using this / these integration(s).

The following integrations are available:

* [MQTT]({{<relref "mqtt.md">}})
* [AWS Simple Notification Service]({{<relref "aws-sns.md"}})
* [Azure Service Bus]({{<relref "azure-service-bus.md">}})
* [Google Cloud Platform Pub/Sub]({{<relref "gcp-pub-sub.md">}})


### Application integrations

Additional to the global enabled integrations, it is also possible to setup
integrations per [application]({{<ref "use/applications.md">}}).

THe following integrations are available:

* [HTTP]({{<relref "http.md">}})
* [InfluxDB]({{<relref "influxdb.md">}})

### Event types

#### Uplink

Contains the data and meta-data for an uplink application payload.
Example:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "rxInfo": [
        {
            "gatewayID": "0303030303030303",          // ID of the receiving gateway
            "name": "rooftop-gateway",                 // name of the receiving gateway
            "time": "2016-11-25T16:24:37.295915988Z",  // time when the package was received (GPS time of gateway, only set when available)
            "rssi": -57,                               // signal strength (dBm)
            "loRaSNR": 10,                             // signal to noise ratio
            "location": {
                "latitude": 52.3740364,  // latitude of the receiving gateway
                "longitude": 4.9144401,  // longitude of the receiving gateway
                "altitude": 10.5,        // altitude of the receiving gateway
            }
        }
    ],
    "txInfo": {
        "frequency": 868100000,  // frequency used for transmission
        "dr": 5                  // data-rate used for transmission
    },
    "adr": false,                  // device ADR status
    "fCnt": 10,                    // frame-counter
    "fPort": 5,                    // FPort
    "data": "...",                 // base64 encoded payload (decrypted)
    "object": {                    // decoded object (when application coded has been configured)
        "temperatureSensor": {"1": 25},
        "humiditySensor": {"1": 32}
    }
}
```

#### Status

Event for battery and margin status received from devices. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "battery": 200,
    "margin": 6,
    "externalPowerSource": false,
    "batteryLevelUnavailable": false,
    "batteryLevel": 75.5
}
```

When configured by the [service-profile]({{<ref "use/service-profiles.md">}})
and when published by the device, this payload contains the device status.

##### battery (deprecated)

* 0 - The end-device is connected to an external power source
* 1...254 - The battery level, 1 being at minimum and 254 being at maximum
* 255 - The end-device was not able to measure the battery level

##### margin

The demodulation signal-to-noise ratio in dB rounded
to the nearest integer valuefor the last successfully received device-status
request by the network-server.

#### Join

Event published when a device joins the network. Please note that this is sent
after the first received uplink (data) frame. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devAddr": "06682ea2",                    // assigned device address
    "devEUI": "0202020202020202"              // device EUI
}
```

#### Ack

Event published on downlink frame acknowledgements. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",             // device EUI
    "acknowledged": true,                     // whether the frame was acknowledged or not (e.g. timeout)
    "fCnt": 12                                // downlink frame-counter
}
```

#### Error

Event published in case of an error related to payload scheduling or handling.
E.g. in case when a payload could not be scheduled as it exceeds the maximum
payload-size. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202"              // device EUI
    "type": "DATA_UP_FCNT",
    "error": "...",
    "fCnt": 123                               // fCnt related to the error (if applicable)
}
```
