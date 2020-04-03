---
title: Devices
menu:
    main:
        parent: use
        weight: 9
description: Manage devices, root-keys and ABP session-keys.
---

# Device management

A device is the end-device connecting to, and communicating over the LoRaWAN<sup>&reg;</sup> network.
ChirpStack Application Server supports both OTAA (over the air activation) and ABP
(activation by personalization) type devices (configured by the selected
[Device Profile]({{<relref "device-profiles.md">}})).

## Create / update

When creating or updating a device, you need to select the
[Device Profile]({{<relref "device-profiles.md">}}) matching the device
capabilities. E.g. the Device Profile defines if the device is of type
OTAA or ABP.

**Note:** In the dropdown, only the Device Profiles are available which
are provisioned on the same [Network Server]({{<relref "network-servers.md">}})
as the [Service Profile]({{<relref "service-profiles.md">}}) which is assigned
to the [Application]({{<relref "applications.md">}}) above the device.

### Tags and variables

Each device can have multiple user-defined tags and variables assigned.
Variables are used for integrations and can for example contain API tokens.
Tags are exposed when ChirpStack Application Server published device events and can be used
to add additional meta-data, e.g. for aggregation.

## Activation

### OTAA devices

After creating a device, you can manage the application-key (and network-key
for LoRaWAN 1.1 devices) under the *Keys (OTAA)* tab. Under the *Activation*
you will see the current device activation (if activated).

### ABP devices

After creating a device, you can ABP activate this device under the
*Activation* tab. You can either enter the *device address*,
*network session key* and *application session key* or generate these.
For LoRaWAN 1.1 devices, the network session key is replaced by
*network session encryption key*, *serving network session integrity key*
and *forwarding network session integrity key*.

## Device provisioning examples

Below you will find provision examples for different devices.

The following example data is used:

{{<highlight text>}}
DevEUI: 0102030405060708
AppKey: 01020304050607080910111213141516
{{< /highlight >}}

In case an App EUI / Join EUI is required, you can set this to a blank value
(`0000000000000000`). This value only becomes relevant when an external
join-server is being used.

### RN2483 / RN2903

Through a serial terminal, use the following commands (terminated by `CR + LF`):

{{<highlight text>}}
mac set deveui 0102030405060708
mac set appeui 0000000000000000
mac set appkey 01020304050607080910111213141516
mac join otaa
{{< /highlight >}}

Any time a `mac save` is done, `mac set devaddr 00000000` should
first be issued. Not doing so will result in the `mac join otaa` command
failing. Per the RN2903 manual, *If this parameter was previously saved to
user EEPROM by issuing the mac save command, after modifying its value, the
mac save command should be called again.*

A `mac save` must be done any time the application key (`mac set appkey`),
application EUI (`mac set appeui`), or the device EUI (`mac set deveui`) are
changed in order to persist these changes. 


### iM880A-L / iM880B-L

Make sure your device is running a recent firmware! These steps were tested
with a device running v1.14. Use [WiMOD LoRaWAN EndDevice Studio](http://www.wireless-solutions.de/products/radiomodules/im880b-l)
for the following actions:

#### To set the DevEUI

1. Go to **Extras** -> **Factory Settings**
2. Click **Set Customer Mode**
3. Enter the **Device EUI** as `0807060504030201`. **Note**: this field
   must be entered as LSB (meaning it is the reverse value as created in
   ChirpStack Application Server)!
4. Click **Set Device EUI**
5. Click **Set Application Mode**

#### To set the AppEUI and AppKey

1. Go back to **LoRaWAN Device**
2. Enter the **Application EUI** as `0000000000000000`. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   ChirpStack Application Server)!
3. Enter the **Application Key** as `01020304050607080910111213141516`.
   **Note:** Opposite to the *Device / Application EUI*, this field must be
   entered as MSB (the same value as set in ChirpStack Application Server).
4. Click **Set Join Parameter**
5. Click **Join Network**

### MultiConnect® mDot™

Through a serial terminal, use the following commands:

{{<highlight text>}}
AT
AT+NJM=1
AT+NI=0,0000000000000000
AT+NK=0,01020304050607080910111213141516
AT&W

ATZ

AT+JOIN
{{< /highlight >}}

For mDot™ we can't modify DevEUI as it's a factory-programmed setting. use the
following command to obtain it:

{{<highlight text>}}
AT+DI?
{{< /highlight >}}

### Generic Arduino LMIC-based devices

1. Make sure that your **Device Profile** has **LoRaWAN MAC version** set to `1.0.2`,
   and **LoRaWAN Regional Parameters revision** set to `A`
2. Install the Arduino LMIC library using the Library Manager in the Arduino IDE
3. Open the example sketch from **Examples -> LMIC-Arduino -> ttn-otaa.ino**
4. Update the sketch with the **Device EUI** as `0807060504030201`. **Note**: this field
   must be entered as LSB (meaning it is the reverse value as created in
   ChirpStack Application Server)!
5. Update the sketch with the **Application EUI** as `0000000000000000`. **Note**: this field
   must be entered as LSB (meaning it is the reverse value as created in
   ChirpStack Application Server)!
6. Update the sketch with the **Application Key** as `01020304050607080910111213141516`.
   **Note:** Opposite to the *Device / Application EUI*, this field must be
   entered as-is (the same value as set in ChirpStack Application Server).
7. Flash the sketch to your device and confirm that the device has been
   activated in the ChirpStack Application Server console and on the Arduino Serial Monitor

### Your device not here?

Please help making this guide complete! Fork the [github.com/brocaar/chirpstack-application-server](https://github.com/brocaar/chirpstack-application-server)
repository, update this page with the actions needed to setup your device
and create a pull-request.
