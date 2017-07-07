---
title: Gateways
menu:
    main:
        parent: use
        weight: 6
---

## Gateway management

An organization is able to manage its own set of gateways. Please note that
this feature might be unavailable when the organization is configured without
gateway support.

That a gateway belongs to a given organization does not mean that the usage 
of a gateway is limited to the organization. Every node in the whole network
will be able to communicate using the gateway. The organization will be
responsible however for managing the gateway details (e.g. name, location)
and will be able to see its statistics.

### Statistics

Gateway statistics are based on the aggregated values sent by the gateway /
packet-forwarder. In case no statistics are visible, it could mean that the
gateway is incorrectly configured.
