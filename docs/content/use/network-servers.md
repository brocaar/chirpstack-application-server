---
title: Network-servers
menu:
    main:
        parent: use
        weight: 3
---

## Network-server management

LoRa App Server is able to connect to one or multiple [LoRa Server](/loraserver/)
network-server instances. Global admin users are able to add new
network-servers to the LoRa App Server installation.

When creating a new network-server, LoRa App Server will create a
routing-profile on the given network-server, containing the `hostname:ip`
of the LoRa App Server installation. In case your LoRa App Server installation
is not reachable on `localhost`, make sure this `hostname:ip` is configured
correctly in your [configuration]({{<ref "install/config.md">}}).
This routing-profile is updated on network-server updates and deleted on
network-server deletes.

**Note:** once a network-server is assigned to a
[service-profile]({{<relref "service-profiles.md">}}) or
[device-profile]({{<relref "device-profiles.md">}}), a network-server can't
be removed before deleting these entities, it will return an error.