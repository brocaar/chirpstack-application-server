---
title: Firmware updates
menu:
  main:
    parent: use
    weight: 13
description: Starting firmware update over the air (FUOTA) jobs.
---

# Firmware update over the air

**Note:** this is an experimental feature, the implementation including
the API might change!

Firmware update over the air (sometimes called FUOTA) makes it possible to
push firmware updates to one or multiple devices, making use of multicast.
It it standardized by the following LoRa<sup>&reg</sup> Alliance specifications:

* LoRaWAN<sup>&reg;</sup> Application Layer Clock Synchronization
* LoRaWAN<sup>&reg;</sup> Fragmented Data Block Transport
* LoRaWAN<sup>&reg;</sup> Remote Multicast Setup

It is important to note that the implementation of this feature by devices
is optional and therefore, unless your device explicitly states that it
implements FUOTA it is safe to assume it does not!

## Starting a firmware update job

Currently ChirpStack Application Server only supports firmware update jobs to single devices.
When navigating to [Devices]({{<relref "devices.md">}}), you will find a
_Firmware_ tab, where you will find the _Create Firmware Update Job_ button.

The following information needs to be provided:

* **Name**: a descriptive name for the update job.
* **Firmware file**: this is the file containing the update (vendor specific).
* **Redundant frames**: the number of extra redundant frames to add to the transmission (more redundancy means that it is more likely a device can recover from packet loss).
* **Unicast timeout**: this is the number of seconds that ChirpStack Application Server will wait for the device to respond to downlink commands.
* **Data-rate**: the used data-rate for the multicast transmission.
* **Frequency**: the frequency used for the multicast transmission.
* **Multicast-group type**: the multicast-group type used.
* **Multicast timeout**: the maximum time the device will enable the configured multicast session (in most cases the device will close the session on receiving the last frame).

## Resources

### ARM Mbed

An example ARM Mbed FUOTA implementation can be found at:
[https://github.com/ARMmbed/mbed-os-example-lorawan-fuota](https://github.com/ARMmbed/mbed-os-example-lorawan-fuota).

To obtain the _Firmware file_ (`xdot-blinky-signed.bin`), you must use the following command:

{{<highlight bash>}}
lorawan-fota-signing-tool sign-binary -b example-firmware/xdot-blinky.bin -o xdot-blinky-signed.bin --output-format bin --override-version
{{</highlight>}}

Refer to the above repository for more information and instructions.
