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
