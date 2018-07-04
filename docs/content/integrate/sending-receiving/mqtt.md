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

```bash
mosquitto_sub -t "application/123/#" -v          # display everything for the given application ID
mosquitto_sub -t "application/123/device/+/rx" -v  # display only the RX payloads for the given application ID
```

**Notes:**

* MQTT topics are case-sensitive
* The `ApplicationID` can be retrieved using the API or from the web-interface,
  this is not the `AppEUI`!

## Receiving

### application/[applicationID]/device/[devEUI]/rx

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

Topic for payloads received from your devices. Example payload:

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

### application/[applicationID]/device/[devEUI]/status

Topic for battery and margin status received from devices. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "battery": 200,
    "margin": 6
}
```

When configured by the [service-profile]({{<ref "use/service-profiles.md">}})
and when published by the device, this payload contains the device status.

#### `battery`

* `0` - The end-device is connected to an external power source
* `1..254` - The battery level, 1 being at minimum and 254 being at maximum
* `255` - The end-device was not able to measure the battery level

#### `margin`

The demodulation signal-to-noise ratio in dB rounded
to the nearest integer valuefor the last successfully received device-status
request by the network-server.

### application/[applicationID]/device/[devEUI]/join

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

Topic for join notifications. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devAddr": "06682ea2",                    // assigned device address
    "DevEUI": "0202020202020202"              // device EUI
}
```

### application/[applicationID]/device/[devEUI]/ack

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

Topic for ACK notifications. Example payload:

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

### application/[applicationID]/device/[devEUI]/error

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

Topic for error notifications. An error might be raised when the downlink
payload size exceeded to max allowed payload size, in case of a MIC error,
... Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "type": "DATA_UP_FCNT",
    "error": "...",
    "fCnt": 123                               // fCnt related to the error (if applicable)
}
```

## Sending

### application/[applicationID]/device/[devEUI]/tx

**Note:** for versions before v1.0.0 `.../device/..` was configured as
`.../node/...`. Please refer to the `application_server.integration.mqtt`
[configuration]({{<ref "install/config.md">}}) for the correct topic.

**Note:** the application ID and DevEUI of the device will be taken from the topic.

Example payload:

```json
{
    "confirmed": true,                        // whether the payload must be sent as confirmed data down or not
    "fPort": 10,                              // FPort to use (must be > 0)
    "data": "...."                            // base64 encoded data (plaintext, will be encrypted by LoRa Server)
    "object": {                               // decoded object (when application coded has been configured)
        "temperatureSensor": {"1": 25},       // when providing the 'object', you can omit 'data'
        "humiditySensor": {"1": 32}
    }
}

```
