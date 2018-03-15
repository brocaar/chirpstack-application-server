---
title: Send / receive data
menu:
    main:
        parent: integrate
        weight: 1
---

# Sending and receiving data

LoRa App Server publishes all the data it receices from the nodes
as JSON over MQTT. To receive data from your node, you therefore 
need to subscribe to its MQTT topic. For debugging, you could use a 
(command-line) tool like `mosquitto_sub` which is part of the 
[Mosquitto](http://mosquitto.org/) MQTT broker.

Use `+` for a single-level wildcard, `#` for a multi-level wildcard.
Examples:

```bash
mosquitto_sub -t "application/123/#" -v          # display everything for the given application ID
mosquitto_sub -t "application/123/node/+/rx" -v  # display only the RX payloads for the given application ID
```

**Notes:**

* MQTT topics are case-sensitive
* The `ApplicationID` can be retrieved using the API or from the web-interface,
  this is not the `AppEUI`!

## Receiving

### application/[applicationID]/node/[devEUI]/rx

Topic for payloads received from your nodes. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "deviceStatusBattery": 200,  // set when available
    "deviceStatusMargin": 6,     // set when available
    "rxInfo": [
        {
            "mac": "0303030303030303",                 // MAC of the receiving gateway
            "name": "rooftop-gateway",                 // name of the receiving gateway
            "latitude": 52.3740364,                    // latitude of the receiving gateway
            "longitude": 4.9144401,                    // longitude of the receiving gateway
            "altitude": 10.5,                          // altitude of the receiving gateway
            "time": "2016-11-25T16:24:37.295915988Z",  // time when the package was received (GPS time of gateway, only set when available)
            "rssi": -57,                               // signal strength (dBm)
            "loRaSNR": 10                              // signal to noise ratio
        }
    ],
    "txInfo": {
        "frequency": 868100000,    // frequency used for transmission
        "dataRate": {
            "modulation": "LORA",  // modulation (LORA or FSK)
            "bandwidth": 250,      // used bandwidth
            "spreadFactor": 5      // used SF (LORA)
            // "bitrate": 50000    // used bitrate (FSK)
        },
        "adr": false,
        "codeRate": "4/6"
    },
    "fCnt": 10,                    // frame-counter
    "fPort": 5,                    // FPort
    "data": "...",                 // base64 encoded payload (decrypted)
    "object": {                    // decoded object (when application coded has been configured)
        "temperatureSensor": {"1": 25},
        "humiditySensor": {"1": 32}
    }
}
```

#### Device-status

When configured by the [service-profile]({{<ref "use/service-profiles.md">}})
and when published by the device, the uplink payload contains the device status.

`deviceStatusBattery`

* `0` - The end-device is connected to an external power source
* `1..254` - The battery level, 1 being at minimum and 254 being at maximum
* `255` - The end-device was not able to measure the battery level

`deviceStatusMargin` is the demodulation signal-to-noise ratio in dB rounded
to the nearest integer valuefor the last successfully received device-status
request by the network-server.

### application/[applicationID]/node/[devEUI]/join

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

### application/[applicationID]/node/[devEUI]/ack

Topic for ACK notifications. Example payload:

```json
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "reference": "abcd1234",                  // the reference given when sending the downlink payload
    "devEUI": "0202020202020202",             // device EUI
    "acknowledged": true,                     // whether the frame was acknowledged or not (e.g. timeout)
    "fCnt": 12                                // downlink frame-counter
}
```

### application/[applicationID]/node/[devEUI]/error

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

### application/[applicationID]/node/[devEUI]/tx

**Note:** the application ID and DevEUI of the node will be taken from the topic.

Example payload:

```json
{
    "reference": "abcd1234",                  // reference which will be used on ack or error (this can be a random string)
    "confirmed": true,                        // whether the payload must be sent as confirmed data down or not
    "fPort": 10,                              // FPort to use (must be > 0)
    "data": "...."                            // base64 encoded data (plaintext, will be encrypted by LoRa Server)
    "object": {                               // decoded object (when application coded has been configured)
        "temperatureSensor": {"1": 25},       // when providing the 'object', you can omit 'data'
        "humiditySensor": {"1": 32}
    }
}

```
