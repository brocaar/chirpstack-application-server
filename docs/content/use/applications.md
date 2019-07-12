---
title: Applications
menu:
    main:
        parent: use
        weight: 7
description: Manage applications, integrations and payload decoders.
---

# Application management

An application is a collection of devices with the same purpose / of the same type.
Think of a weather station collecting data at different locations for example.

When creating an application, you need to select the [Service-profile]({{<relref "service-profiles.md">}})
which will be used for the devices created under this application. Note that
once a service-profile has been selected, it can't be changed.

## Payload codecs

The payload codec options have moved to the [Device Profile]({{<relref "device-profiles.md">}}).
For backward compatibility, existing codec configuration on the application is still accessible
and functional, but this will be removed fully in the next major release update.

## Integrations

For documentation on the available integrations, please refer to
[sending and receiving](/lora-app-server/integrate/sending-receiving/).

## Devices

Multiple [devices]({{<relref "devices.md">}}) can be added to the application.
