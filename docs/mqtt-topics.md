# MQTT topics

To receive data from your node, you need to subscribe to its MQTT topic.
For debugging, you can use a (command-line) tool like ``mosquitto_sub``
which is part of the [Mosquitto](http://mosquitto.org/) MQTT broker.

Use ``+`` for a single-level wildcard, ``#`` for a multi-level wildcard.
Examples:

```bash
mosquitto_sub -t "application/temperature-sensor/#" -v          # display everything for the given application
mosquitto_sub -t "application/temperature-sensor/node/+/rx" -v  # display only the RX payloads for the given application
```

!!! info
	Note that the MQTT topics are case-sensitive

## Receiving

### application/[applicationName]/node/[nodeName]/rx

Topic for payloads received from your nodes. Example payload:

```json
{
	"applicationName": "temperature-sensor",
	"nodeName": "garden-sensor",
	"devEUI": "0202020202020202",
    "rxInfo": [
        {
            "mac": "0303030303030303",                 // MAC of the receiving gateway
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
    "data": "..."                  // base64 encoded payload (decrypted)
}
```

### application/[applicationName]/node/[nodeName]/join

Topic for join notifications. Example payload:

```json
{
	"applicationName": "temperature-sensor",
	"nodeName": "garden-sensor",
    "devAddr": "06682ea2",                    // assigned device address
    "DevEUI": "0202020202020202"              // device EUI
}
```

### application/[applicationName]/node/[nodeName]/ack

Topic for ACK notifications. Example payload:

```json
{
	"applicationName": "temperature-sensor",
	"nodeName": "garden-sensor",
    "reference": "abcd1234",                  // the reference given when sending the downlink payload
    "devEUI": "0202020202020202"              // device EUI
}
```

### application/[applicationName]/node/[nodeName]/error

Topic for error notifications. An error might be raised when the downlink
payload size exceeded to max allowed payload size. Please see the LoRaWAN
specification for the max allowed payload size for your region. Example:

```json
{
    "reference": "abcd1234",    // the reference given when sending the downlink payload
    "message": "error message"  // the content of the error message
}
```

## Sending

### application/[applicationName]/node/[nodeName]/tx

!!! info
	The name of the application and node will be taken from the MQTT topic.

Example payload:

```json
{
    "reference": "abcd1234",                  // reference which will be used on ack or error (this can be a random string)
    "confirmed": true,                        // whether the payload must be sent as confirmed data down or not
    "fPort": 10,                              // FPort to use (must be > 0)
    "data": "...."                            // base64 encoded data (plaintext, will be encrypted by LoRa Server)
}

```
