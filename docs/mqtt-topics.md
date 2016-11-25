# MQTT topics

To receive data from your node, you need to subscribe to its MQTT topic.
For debugging, you can use a (command-line) tool like ``mosquitto_sub``
which is part of the [Mosquitto](http://mosquitto.org/) MQTT broker.

Use ``+`` for a single-level wildcard, ``#`` for a multi-level wildcard.
Examples:

```bash
mosquitto_sub -t "application/0101010101010101/#" -v          # display everything for the given application
mosquitto_sub -t "application/0101010101010101/node/+/rx" -v  # display only the RX payloads for the given application
```

## Receiving

### application/[AppEUI]/node/[DevEUI]/rx

Topic for payloads received from your nodes. Example payload:

```json
{
    "devEUI": "0202020202020202",  // device EUI
    "rxInfo": [
        {
            "mac": "0303030303030303",                 // MAC of the receiving gateway
            "time": "2016-11-25T16:24:37.295915988Z",  // time when the package was received (GPS time of gateway, only set when available)
            "rssi": -57,                               // signal strength (dBm)
            "loRaSNR": 10                              // signal to noise ratio
        }
    ],
    "fCnt": 10,                    // frame-counter
    "fPort": 5,                    // FPort
    "data": "..."                  // base64 encoded payload (decrypted)
}
```

### application/[AppEUI]/node/[DevEUI]/join

Topic for join notifications. Example payload:

```json
{
    "devAddr": "06682ea2",        // device address
    "DevEUI": "0202020202020202"  // device EUI
}
```

### application/[AppEUI]/node/[DevEUI]/ack

Topic for ACK notifications. Example payload:

```json
{
    "reference": "abcd1234",      // the reference given when sending the downlink payload
    "devEUI": "0202020202020202"  // device EUI
}
```

### application/[AppEUI]/node/[DevEUI]/error

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

### application/[AppEUI]/node/[DevEUI]/tx

Example payload:

```json
{
    "reference": "abcd1234",       // reference given by the application, will be used on error
    "confirmed": true,             // whether the payload must be sent as confirmed data down or not
    "devEUI": "0202020202020202",  // the device to sent the data to
    "fPort": 10,                   // FPort to use
    "data": "...."                 // base64 encoded data (plaintext, will be encrypted by LoRa Server)
}

```
