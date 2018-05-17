---
title: Applications
menu:
    main:
        parent: use
        weight: 7
---

# Application management

An application is a collection of devices with the same purpose / of the same type.
Think of a weather station collecting data at different locations for example.

When creating an application, you need to select the [Service-profile]({{<relref "service-profiles.md">}})
which will be used for the devices created under this application. Note that
once a service-profile has been selected, it can't be changed.

An application can be configured to decode the received uplink payloads from
bytes to a meaningful data object, and to encode downlink objects to bytes.

## Payload codecs

**Note:** the raw `base64` encoded payload will always be available, even when
a codec has been configured.

### Cayenne LPP

When selecting the Cayenne LPP codec, LoRa App Server will decode and encode
following the [Cayenne Low Power Payload](https://mydevices.com/cayenne/docs/lora/)
specification.

### Custom JavaScript codec functions

When selecting the Custom JavaScript codec functions option, you can write your
own (JavaScript) functions to decode an array of bytes to a JavaScript object
and encode a JavaScript object to an array of bytes.

#### Decoder function skeleton

```js
// Decode decodes an array of bytes into an object.
//  - fPort contains the LoRaWAN fPort number
//  - bytes is an array of bytes, e.g. [225, 230, 255, 0]
// The function must return an object, e.g. {"temperature": 22.5}
function Decode(fPort, bytes) {
  return {};
}
```

#### Encoder function skeleton

```js
// Encode encodes the given object into an array of bytes.
//  - fPort contains the LoRaWAN fPort number
//  - obj is an object, e.g. {"temperature": 22.5}
// The function must return an array of bytes, e.g. [225, 230, 255, 0]
function Encode(fPort, obj) {
  return [];
}
```

## Integrations

For documentation on the available integrations, please refer to
[sending and receiving](/lora-app-server/integrate/sending-receiving/).

## Devices

Multiple [devices]({{<relref "devices.md">}}) can be added to the application.