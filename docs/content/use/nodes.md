---
title: Nodes
menu:
    main:
        parent: use
        weight: 5
---

## Node management

A node is the device connecting to and communicating over the LoRaWAN network.
LoRa App Server supports both OTAA (over the air activation) and ABP
(activation by personalization) type nodes.


### OTAA

In case of an OTAA node, you need the following details to add a new node:

* Device EUI
* Application EUI
* Application Key

### ABP

In case of an ABP node, you need the following details to create the node:

* Device EUI
* Application EUI

After creating the node, click on the ABP activation button to enter the:

* Device Address
* Network Session Key
* Application Session Key

### Settings

When the *Use application settings* checkbox is ticked, the node will inherit
the application settings. In this case all settings are visible, but not
editable. Uncheck this checkbox to make any modifications specific to the node.

#### Receive window

Based on this configuration, the first or second receive-window will be used
for downlink communication with the node.

#### Relax frame-counter mode

As LoRa Server will keep track of the frame-counters (for security reasons),
this could be an issue for ABP nodes as a power-cycle means it will reset
the frame-counters. To work around this issue, the *Enable relax frame-counter*
mode can be activated. In this case LoRa Server will reset all frame-counters
to `0` when it detects that a node was reset. Please note that this introduces
a security risk.

#### Adaptive data-rate

To enable ADR support (from the network-side, please note that you need to
enable ADR on your node too), you must set the:

* ADR interval
* Installation margin

### Node provisioning

After setting up a node in LoRa App Server, you need to
provision your node with the chosen AppEUI, DevEUI and AppKey.
This document will describe this process for different types of nodes.

The following example data is used:

```
DevEUI: 0102030405060708
AppEUI: 0807060504030201
AppKey: 01020304050607080910111213141516
```

#### RN2483 / RN2903

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


#### iM880A-L / iM880B-L

Make sure your node is running a recent firmware! These steps were tested
with a node running v1.14. Use [WiMOD LoRaWAN EndNode Studio](http://www.wireless-solutions.de/products/radiomodules/im880b-l)
for the following actions:

##### To set the DevEUI

1. Go to **Extras** -> **Factory Settings**
2. Click **Set Customer Mode**
3. Enter the **Device EUI** as ``0807060504030201``. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   LoRa Server)!
4. Click **Set Device EUI**
5. Click **Set Application Mode**

##### To set the AppEUI and AppKey

1. Go back to **LoRaWAN Node**
2. Enter the **Application EUI** as ``0102030405060708``. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   LoRa Server)!
3. Enter the **Application Key** as ``01020304050607080910111213141516``.
   **Note:** Opposite to the *Device / Application EUI*, this field must be
   entered as-is (the same value as set in LoRa Server).
4. Click **Set Join Parameter**
5. Click **Join Network**

#### MultiConnect® mDot™

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

#### Your node not here?

Please help making this guide complete! Fork the [github.com/brocaar/lora-app-server](https://github.com/brocaar/lora-app-server)
repository, update this page with the actions needed to setup your node
and create a pull-request :-)
