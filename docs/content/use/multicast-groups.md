---
title: Multicast Groups
menu:
    main:
        parent: use
        weight: 8
description: Manage multicast-groups for sending downlink frames to a group of devices (the multicast-group).
---

# Multicast Groups

By creating a multicast-group, it is possible to send a single downlink payload
to a group of devices (the multicast-group). All these devices share the same
multicast-address, session-keys and frame-counter.

After creating a multicast-group, it is possible to assign devices to the group.
Please note that the device must already created (see [Devices]({{<relref "devices.md">}})).
Only devices that share the same [Service Profile]({{<relref "service-profiles.md">}})
as the multicast-group can be added.

## Provisioning of the device

The provisioning of the multicast-group on the device happens out-of-band.
This means that after adding a device to a multicast-group, you must also
configure the device with the multicast-address, session-keys etc...

## Sending data

Sending data to the multicast-group happens using the [gRPC]({{<ref "/integrate/grpc.md">}})
or [RESTful JSON]({{<ref "/integrate/rest.md">}}) API.
