---
title: Network Servers
menu:
    main:
        parent: use
        weight: 3
description: Manage the connected Network Servers (ChirpStack Network Server instances), supporting multiple regions.
---

# Network Server management

## Network Servers

ChirpStack Application Server is able to connect to one or multiple [ChirpStack Network Server](/Network Server/)
instances. Global admin users are able to add new
Network Servers to the ChirpStack Application Server installation.

**Note:** once a Network Server is assigned to a
[Service Profile]({{<relref "service-profiles.md">}}) or
[Device Profile]({{<relref "device-profiles.md">}}), a Network Server can't
be removed before deleting these entities, it will return an error.

## Routing-profile

When creating a new Network Server, ChirpStack Application Server will create a
Routing Profile on the given Network Server, containing the `hostname:ip`
of the ChirpStack Application Server installation. In case your ChirpStack Application Server installation
is not reachable on `localhost`, make sure this `hostname:ip` is configured
correctly in your [Configuration]({{<ref "install/config.md">}}).
This Routing Profile is updated on Network Server updates and deleted on
Network Server deletes.

## TLS certificates

Depending the configuration of ChirpStack Network Server and ChirpStack Application Server, you must enter
the CA and client certificates in order to let ChirpStack Application Server connect to
ChirpStack Network Server and in order to let ChirpStack Network Server connect to ChirpStack Application Server
(see *Routing-profile* above).

Note that for security reasons, the *TLS key* content is not displayed
when editing an existing Network Server. Re-submitting does not clear the
stored TLS key when left blank! The *TLS key* content will only be cleared
internally when submitting a form with an empty *TLS certificate* value.

### ChirpStack Network Server API is using TLS

You must enter the CA and TLS certificate fields under
**Certificates for ChirpStack Application Server to ChirpStack Network Server connection**.

See also [ChirpStack Network Server configuration](https://www.chirpstack.io/network-server/install/config/).

### ChirpStack Application Server API is using TLS

You must enter the CA and TLS certificate fields under
**Certificates for ChirpStack Network Server to ChirpStack Application Server connection**.

See also [ChirpStack Application Server configuration]({{<ref "install/config.md">}}).
