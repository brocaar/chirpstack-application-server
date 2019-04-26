---
title: Device-profiles
menu:
    main:
        parent: use
        weight: 6
toc: false
description: Manage device-profiles, containing the device capabilities and boot parameters.
---

# Device-profile management

A device-profile defines the device capabilities and boot parameters
that are needed by the network-server for setting the LoRaWAN radio
access service. These information elements shall be provided by the
end-device manufacturer.

When creating a device-profile, LoRa App Server will create the actual
profile on the selected network-server, and will keep a reference record
so it knows to which organization it belongs.

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
and encode a JavaScript object to an array of bytes. Package [otto](https://github.com/robertkrimen/otto), 
which targets ES5, is used as a JavaScript interpreter, so ES6 features (e.g. Typed Arrays) are not supported.

#### Decoder function skeleton

{{<highlight js>}}
// Decode decodes an array of bytes into an object.
//  - fPort contains the LoRaWAN fPort number
//  - bytes is an array of bytes, e.g. [225, 230, 255, 0]
// The function must return an object, e.g. {"temperature": 22.5}
function Decode(fPort, bytes) {
  return {};
}
{{< /highlight >}}

#### Encoder function skeleton

{{<highlight js>}}
// Encode encodes the given object into an array of bytes.
//  - fPort contains the LoRaWAN fPort number
//  - obj is an object, e.g. {"temperature": 22.5}
// The function must return an array of bytes, e.g. [225, 230, 255, 0]
function Encode(fPort, obj) {
  return [];
}
{{< /highlight >}}

## Fields / options

The following fields are described by the
[LoRaWAN Backend Interfaces specification](https://www.lora-alliance.org/lorawan-for-developers).
Fields marked with an **X** are implemented by LoRa (App) Server.

- [X] **SupportsClassB** End-Device supports Class B
- [X] **ClassBTimeout** Maximum delay for the End-Device to answer a MAC request or a confirmed DL frame (mandatory if class B mode supported)
- [X] **PingSlotPeriod** Mandatory if class B mode supported
- [X] **PingSlotDR** Mandatory if class B mode supported
- [X] **PingSlotFreq** Mandatory if class B mode supported
- [X] **SupportsClassC** End-Device supports Class C
- [X] **ClassCTimeout** Maximum delay for the End-Device to answer a MAC request or a confirmed DL frame (mandatory if class C mode supported)
- [X] **MACVersion** Version of the LoRaWAN supported by the End-Device
- [X] **RegParamsRevision** Revision of the Regional Parameters document supported by the End-Device
- [X] **SupportsJoin** End-Device supports Join (OTAA) or not (ABP)
- [X] **RXDelay1** Class A RX1 delay (mandatory for ABP)
- [X] **RXDROffset1** RX1 data rate offset (mandatory for ABP)
- [X] **RXDataRate2** RX2 data rate (mandatory for ABP)
- [X] **RXFreq2** RX2 channel frequency (mandatory for ABP)
- [X] **FactoryPresetFreqs** List of factory-preset frequencies (mandatory for ABP)
- [X] **MaxEIRP** Maximum EIRP supported by the End-Device
- [ ] **MaxDutyCycle** Maximum duty cycle supported by the End-Device
- [X] **RFRegion** RF region name (automatically set by LoRa Server)
- [ ] **Supports32bitFCnt** End-Device uses 32bit FCnt (mandatory for LoRaWAN 1.0 End-Device) (always set to `true`)

