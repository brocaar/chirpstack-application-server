---
title: Gateways
menu:
    main:
        parent: use
        weight: 9
---

# Gateway management

An organization is able to manage its own set of gateways. Please note that
this feature might be unavailable when the organization is configured without
gateway support.

That a gateway belongs to a given organization does not mean that the usage 
of a gateway is limited to the organization. Every node in the whole network
will be able to communicate using the gateway. The organization will be
responsible however for managing the gateway details (e.g. name, location)
and will be able to see its statistics.

## Statistics

Gateway statistics are based on the aggregated values sent by the gateway /
packet-forwarder. In case no statistics are visible, it could mean that the
gateway is incorrectly configured.

## Channel-management

The channel-configuration of the gateways within the network can be managed
by LoRa Server, using the [LoRa Gateway Config](/lora-gateway-config/)
component. This is not required and is an optional feature.

### Channel configuration

Global admin users are able to create channel configurations from the
top navbar in the web-interface (or by using the API). Channel
configuration consists of a name (used as an identifier) and the
enabled channel numbers (as defined by the [LoRaWAN Regional Parameters](https://www.lora-alliance.org/lorawan-for-developers)
specification). Multiple configurations can be created to distribute
different channels across different gateways. 

The created channel-configuration can be used by everybody who has
permission to create / modify gateways to assign these configurations to
gateways.

#### Extra channels

For some ISM bands it is possible to create extra channels. This can be
used for bands that allow the CFList option (up to 5 extra channels of
125 kHz using spread-factors 7-12) or to configure the single spread-factor
LoRa (250 or 500 kHz) and FSK channel.

**Notes:**

* For some bands (e.g. the US ISM band) the single spread-factor is already
  defined by the LoRaWAN Regional Parameters specification.
* When defining extra channels, make sure these channels fit within the
  bandwidth of the radios used by your gateway. For the SX1257 the available
  bandwidth is:
  	* 1.1 MHz for 500kHz channels
	* 1 MHz for 250kHz channels
	* 0.925 MHz for 125kHz channels


