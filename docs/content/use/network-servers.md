---
title: Network-servers
menu:
    main:
        parent: use
        weight: 3
---

# Network-server management

## Network-servers

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

## Gateway-profiles

Once a network-server has been created, it is possible to provision one or more
gateway-profiles on this network-server (available under the *Gateway-profiles*
tab). When adding a gateway, it is then possible to select one of these
gateway-profiles to make sure the gateway configuration is in sync with the
channels used by the network.

Please note that this must also be configured
in your [LoRa Gateway Bridge configuration](/lora-gateway-bridge/install/config/).

**Important:** changing the channel-plan of the gateway does not change the
channels used by the devices. For a correct setup, the gateway-profile(s)
must at least support the channels used by your devices. The channels that must
be used by your devices must be configured in your
[LoRa Server configuration](/loraserver/install/config/).

### Enabled channels

Enter the list of default channels that you would like to use for this
gateway-profile. Note that these are the channels that are specified by the
[LoRaWAN Regional Parameters](https://www.lora-alliance.org/lorawan-for-developers).

### Extra channels

If allowed by your LoRaWAN region, you can use *Add extra channel* to configure
additional channels that are not defined by the LoRaWAN Regional Parameters.

### Hardware limitations

This feature is limited to 8-channel gateways (currently) and assumes that
channels can be distributed over two radios. When defining a channel-plan,
please keep in mind that the channels fit within the bandwidth of two radios.

The bandwidth of each radio depends on the bandwidth of the assigned channels:

* 500kHz channel = 1.1MHz radio bandwidth
* 250kHz channel = 1Mhz radio bandwidth
* 125kHz channel = 0.925MHz radio bandwidth
