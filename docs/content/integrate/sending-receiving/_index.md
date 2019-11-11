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

Global integrations are configured throught the [chirpstack-application-server.toml]({{<ref "install/config.md">}})
configuration file and are globally enabled. This means that for every application configured,
the application will be published using this / these integration(s).

The following integrations are available:

* [MQTT]({{<relref "mqtt.md">}})
* [PostgreSQL]({{<relref "postgresql.md">}})
* [AWS Simple Notification Service]({{<relref "aws-sns.md">}})
* [Azure Service Bus]({{<relref "azure-service-bus.md">}})
* [Google Cloud Platform Pub/Sub]({{<relref "gcp-pub-sub.md">}})


### Application integrations

Additional to the global enabled integrations, it is also possible to setup
integrations per [application]({{<ref "use/applications.md">}}).

The following integrations are available:

* [HTTP]({{<relref "http.md">}})
* [InfluxDB]({{<relref "influxdb.md">}})
* [ThingsBoard]({{<relref "thingsboard.md">}})

### Event types

Event payloads can be encoded into different payload encodings, using the
`marshaler` configuration option under `[application_server.integrations]`
in the [Configuration]({{<ref "install/config.md">}}) file. For the [Protobuf](https://developers.google.com/protocol-buffers/)
message definitions, please refer to [integration.proto](https://github.com/brocaar/chirpstack-api/blob/master/protobuf/as/integration/integration.proto).

* JSON (v3): ChirpStack Application Server v3 JSON format, will be removed in the next major ChirpStack Application Server version
* JSON: Protobuf based JSON format
* Protobuf: Protobuf binary encoding

**Note:** The Protocol Buffers [JSON Mapping](https://developers.google.com/protocol-buffers/docs/proto3#json)
defines that bytes must be encoded as base64 strings. This applies to the
devEUI field for example. When re-encoding this filed to HEX encoding, you
will find the expected devEUI string.

#### Uplink

Contains the data and meta-data for an uplink application payload.

##### JSON v3

{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "rxInfo": [
        {
            "gatewayID": "0303030303030303",
            "name": "rooftop-gateway",
            "time": "2016-11-25T16:24:37.295915988Z",
            "rssi": -57,
            "loRaSNR": 10,
            "location": {
                "latitude": 52.3740364,
                "longitude": 4.9144401,
                "altitude": 10.5
            }
        }
    ],
    "txInfo": {
        "frequency": 868100000,
        "dr": 5
    },
    "adr": false,
    "fCnt": 10,
    "fPort": 5,
    "data": "...",
    "object": {
        "temperatureSensor": {"1": 25},
        "humiditySensor": {"1": 32}
    },
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf JSON

{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "AgICAgICAgI=",
    "rxInfo": [
        {
            "gatewayID": "AwMDAwMDAwM=",
            "time": "2019-11-08T13:59:25.048445Z",
            "timeSinceGPSEpoch": null,
            "rssi": -48,
            "loRaSNR": 9,
            "channel": 5,
            "rfChain": 0,
            "board": 0,
            "antenna": 0,
            "location": {
                "latitude": 52.3740364,
                "longitude": 4.9144401,
                "altitude": 10.5
            },
            "fineTimestampType": "NONE",
            "context": "9u/uvA==",
            "uplinkID": "jhMh8Gq6RAOChSKbi83RHQ=="
        }
    ],
    "txInfo": {
        "frequency": 868100000,
        "modulation": "LORA",
        "loRaModulationInfo": {
            "bandwidth": 125,
            "spreadingFactor": 11,
            "codeRate": "4/5",
            "polarizationInversion": false
        }
    },
    "adr": true,
    "dr": 1,
    "fCnt": 10,
    "fPort": 5,
    "data": "...",
    "objectJSON": "{\"temperatureSensor\":25,\"humiditySensor\":32}",
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf

This message is defined by the `UplinkEvent` Protobuf message.

#### Status

Event for battery and margin status received from devices.

The interval in which the Network Server will request the device-status is
configured by the [service-profile]({{<ref "use/service-profiles.md">}}).

##### JSON v3


{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "battery": 200,
    "margin": 6,
    "externalPowerSource": false,
    "batteryLevelUnavailable": false,
    "batteryLevel": 75.5,
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf JSON

{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "AgICAgICAgI=",
    "margin": 6,
    "externalPowerSource": false,
    "batteryLevelUnavailable": false,
    "batteryLevel": 75.5,
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf

This message is defined by the `StatusEvent` Protobuf message.

#### Join

Event published when a device joins the network. Please note that this is sent
after the first received uplink (data) frame.

##### JSON v3


{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devAddr": "06682ea2",
    "devEUI": "0202020202020202",
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf JSON

{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "AgICAgICAgI=",
    "devAddr": "AFE5Qg==",
    "rxInfo": [
        {
            "gatewayID": "AwMDAwMDAwM=",
            "time": "2019-11-08T13:59:25.048445Z",
            "timeSinceGPSEpoch": null,
            "rssi": -48,
            "loRaSNR": 9,
            "channel": 5,
            "rfChain": 0,
            "board": 0,
            "antenna": 0,
            "location": {
                "latitude": 52.3740364,
                "longitude": 4.9144401,
                "altitude": 10.5
            },
            "fineTimestampType": "NONE",
            "context": "9u/uvA==",
            "uplinkID": "jhMh8Gq6RAOChSKbi83RHQ=="
        }
    ],
    "txInfo": {
        "frequency": 868100000,
        "modulation": "LORA",
        "loRaModulationInfo": {
            "bandwidth": 125,
            "spreadingFactor": 11,
            "codeRate": "4/5",
            "polarizationInversion": false
        }
    },
    "dr": 1,
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf

This message is defined by the `JoinEvent` Protobuf message.

#### Ack

Event published on downlink frame acknowledgements.

##### JSON v3


{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "acknowledged": true,
    "fCnt": 12,
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf JSON

{{<highlight json>}}
{

    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "AgICAgICAgI=",
    "rxInfo": [
        {
            "gatewayID": "AwMDAwMDAwM=",
            "time": "2019-11-08T13:59:25.048445Z",
            "timeSinceGPSEpoch": null,
            "rssi": -48,
            "loRaSNR": 9,
            "channel": 5,
            "rfChain": 0,
            "board": 0,
            "antenna": 0,
            "location": {
                "latitude": 52.3740364,
                "longitude": 4.9144401,
                "altitude": 10.5
            },
            "fineTimestampType": "NONE",
            "context": "9u/uvA==",
            "uplinkID": "jhMh8Gq6RAOChSKbi83RHQ=="
        }
    ],
    "txInfo": {
        "frequency": 868100000,
        "modulation": "LORA",
        "loRaModulationInfo": {
            "bandwidth": 125,
            "spreadingFactor": 11,
            "codeRate": "4/5",
            "polarizationInversion": false
        }
    },
	"acknowledged": true,
	"fCnt": 15,
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf

This message is defined by the `AckEvent` Protobuf message.


#### Error

Event published in case of an error related to payload scheduling or handling.
E.g. in case when a payload could not be scheduled as it exceeds the maximum
payload-size.

##### JSON v3


{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "0202020202020202",
    "type": "DATA_UP_FCNT",
    "error": "...",
    "fCnt": 123,
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf JSON

{{<highlight json>}}
{
    "applicationID": "123",
    "applicationName": "temperature-sensor",
    "deviceName": "garden-sensor",
    "devEUI": "AgICAgICAgI=",
	"type": "UPLINK_CODEC",
	"error": "...",
    "tags": {
        "key": "value"
    }
}
{{</highlight>}}

##### Protobuf

This message is defined by the `ErrorEvent` Protobuf message.
