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

## Gateway-profiles

When assigning a gateway-profile to a gateway, [LoRa Server](/loraserver/)
will send a configuration update to the gateway when its configuration
is out-of-date. Gateway-profiles can be configured by clicking on 
[Network servers]({{<relref "network-servers.md" >}}) in the navbar.
Note that when a gateway-profile is assigned to a gateway, you must also
configure the [LoRa Gateway Bridge configuration](/lora-gateway-bridge/install/config/)
in order to handle configuration updates.
