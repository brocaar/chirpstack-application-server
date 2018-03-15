---
title: Devices
menu:
    main:
        parent: use
        weight: 8
---

# Device management

A device is the end-device connecting to, and communicating over the LoRaWAN network.
LoRa App Server supports both OTAA (over the air activation) and ABP
(activation by personalization) type devices (configured by the selected
[device-profile]({{<relref "device-profiles.md">}})).

## Create / update

When creating or updating a device, you need to select the
[device-profile]({{<relref "device-profiles.md">}}) matching the device
capabilities. E.g. the device-profile defines if the device is of type
OTAA or ABP.

**Note:** In the dropdown, only the device-profiles are available which
are provisioned on the same [network-server]({{<relref "network-servers.md">}})
as the [service-profile]({{<relref "service-profiles.md">}}) which is assigned
to the [application]({{<relref "applications.md">}}) above the device.

## Activation

### OTAA devices

After creating a device, you can manage the application-key under
the *Device keys (OTAA)* tab. Under the *Device activation* you will see the
current device activation (if activated).

### ABP devices

After creating a device, you can ABP activate this device under the
*Activate device (ABP)* tab. You can either enter the *device address*,
*network session key* and *application session key* or generate these.

After the ABP device has been activated, the current activation can be seen
under the *Device activation* tab.

## Device provisioning

After setting up a device in LoRa App Server, you need to
provision your device with the chosen AppEUI, DevEUI and AppKey.
This document will describe this process for different types of devices.

The following example data is used:

```
DevEUI: 0102030405060708
AppEUI: 0807060504030201
AppKey: 01020304050607080910111213141516
```

### RN2483 / RN2903

Through a serial terminal, use the following commands (terminated by `CR + LF`):

```
mac set deveui 0102030405060708
mac set appeui 0807060504030201
mac set appkey 01020304050607080910111213141516
mac join otaa
```

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
3. Enter the **Device EUI** as ``0807060504030201``. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   LoRa Server)!
4. Click **Set Device EUI**
5. Click **Set Application Mode**

#### To set the AppEUI and AppKey

1. Go back to **LoRaWAN Device**
2. Enter the **Application EUI** as ``0102030405060708``. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   LoRa Server)!
3. Enter the **Application Key** as ``01020304050607080910111213141516``.
   **Note:** Opposite to the *Device / Application EUI*, this field must be
   entered as-is (the same value as set in LoRa Server).
4. Click **Set Join Parameter**
5. Click **Join Network**

### MultiConnect® mDot™

Through a serial terminal, use the following commands:

```
AT
AT+NJM=1
AT+NI=0,0807060504030201
AT+NK=0,01020304050607080910111213141516
AT&W

ATZ

AT+JOIN
```

For mDot™ we can't modify DevEUI as it's a factory-programmed setting. use the
following command to obtain it:

```
AT+DI?
```

### Your device not here?

Please help making this guide complete! Fork the [github.com/brocaar/lora-app-server](https://github.com/brocaar/lora-app-server)
repository, update this page with the actions needed to setup your device
and create a pull-request :-)
