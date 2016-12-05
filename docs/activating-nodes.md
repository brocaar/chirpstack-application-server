# Activating nodes

After setting up a node in LoRa App Server
(see [getting started](getting-started.md) for more details), you need to
provision your node with the chosen AppEUI, DevEUI and AppKey.
This document will describe this process for different types of nodes.

The following example data is used:

```
DevEUI: 0102030405060708
AppEUI: 0807060504030201
AppKey: 01020304050607080910111213141516
```

## RN2483

Through a serial terminal, use the following commands:

```
mac set deveui 0102030405060708
mac set appeui 0807060504030201
mac set appkey 01020304050607080910111213141516
mac join otaa
```

## RN2903

```
mac set deveui 0102030405060708
mac set appeui 0807060504030201
mac set appkey 01020304050607080910111213141516
mac join otaa
```

Some additional notes:

All unused channels should be disabled with `mac set ch status <channel_id> off`,
where `channel_id` is `8..71`, assuming you're using channels `0..7`.
This should be followed by a `mac save` command to save the channel status
to EEPROM.

Secondly, any time a `mac save` is done, `mac set devaddr 00000000` should
first be issued. Not doing so will result in the `mac join otaa` command
failing. Per the RN2903 manual, *If this parameter was previously saved to
user EEPROM by issuing the mac save command, after modifying its value, the
mac save command should be called again.*

A `mac save` must be done any time the application key (`mac set appkey`),
application EUI (`mac set appeui`), or the device EUI (`mac set deveui`) are
changed in order to persist these changes. 

Anyone who is this far along probably has figured out that commands to the
RN2483 / RN2903 must be terminated with `CR + LF`.

## iM880A-L / iM880B-L

Make sure your node is running a recent firmware! These steps were tested
with a node running v1.14. Use [WiMOD LoRaWAN EndNode Studio](http://www.wireless-solutions.de/products/radiomodules/im880b-l)
for the following actions:

### To set the DevEUI

1. Go to **Extras** -> **Factory Settings**
2. Click **Set Customer Mode**
3. Enter the **Device EUI** as ``0807060504030201``. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   LoRa Server)!
4. Click **Set Device EUI**
5. Click **Set Application Mode**

### To set the AppEUI and AppKey

1. Go back to **LoRaWAN Node**
2. Enter the **Application EUI** as ``0102030405060708``. **Note**: this field
   must be entered as LSBF (meaning it is the reverse value as created in
   LoRa Server)!
3. Enter the **Application Key** as ``01020304050607080910111213141516``.
   **Note:** Opposite to the *Device / Application EUI*, this field must be
   entered as-is (the same value as set in LoRa Server).
4. Click **Set Join Parameter**
5. Click **Join Network**

## MultiConnect® mDot™

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

!!! info "Your node not here?"
    Please help making this guide complete! Fork the [github.com/brocaar/loraserver](https://github.com/brocaar/loraserver)
    repository, update this page with the actions needed to setup your node
    and create a pull-request :-)
