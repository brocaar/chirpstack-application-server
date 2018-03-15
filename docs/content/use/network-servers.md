---
title: Network-servers
menu:
    main:
        parent: use
        weight: 3
---

# Network-server management

LoRa App Server is able to connect to one or multiple [LoRa Server](/loraserver/)
network-server instances. Global admin users are able to add new
network-servers to the LoRa App Server installation.

**Note:** once a network-server is assigned to a
[service-profile]({{<relref "service-profiles.md">}}) or
[device-profile]({{<relref "device-profiles.md">}}), a network-server can't
be removed before deleting these entities, it will return an error.

## Routing-profile

When creating a new network-server, LoRa App Server will create a
routing-profile on the given network-server, containing the `hostname:ip`
of the LoRa App Server installation. In case your LoRa App Server installation
is not reachable on `localhost`, make sure this `hostname:ip` is configured
correctly in your [configuration]({{<ref "install/config.md">}}).
This routing-profile is updated on network-server updates and deleted on
network-server deletes.

## TLS certificates

Depending the configuration of LoRa Server and LoRa App Server, you must enter
the CA and client certificates in order to let LoRa App Server connect to
LoRa Server and in order to let LoRa Server connect to LoRa App Server
(see *Routing-profile* above).

Note that for security reasons, the *TLS key* content is not displayed
when editing an existing network-server. Re-submitting does not clear the
stored TLS key when left blank! The *TLS key* content will only be cleared
internally when submitting a form with an empty *TLS certificate* value.

### LoRa Server API is using TLS

You must enter the CA and TLS certificate fields under
**Certificates for LoRa App Server to LoRa Server connection**.

See also [LoRa Server configuration](https://docs.loraserver.io/loraserver/install/config/).

### LoRa App Server API is using TLS

You must enter the CA and TLS certificate fields under
**Certificates for LoRa Server to LoRa App Server connection**.

See also [LoRa App Server configuration]({{<ref "install/config.md">}}).
