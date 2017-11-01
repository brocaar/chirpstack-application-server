---
title: Applications
menu:
    main:
        parent: use
        weight: 7
---

## Application management

An application is a collection of devices with the same purpose / of the same type.
Think of a weather station collecting data at different locations for example.

When creating an application, you need to select the [Service-profile]({{<relref "service-profiles.md">}})
which will be used for the devices created under this application. Note that
once a service-profile has been selected, it can't be changed.

An application has:

* Integrations
* Devices

### Integrations

By default all data is published to a MQTT broker, see also
[Sending and receiving data]({{<ref "integrate/data.md">}}). However additional
integrations can be setup. See [Integrations]({{<ref "integrate/integrations.md">}})
for more information.

### Devices

Multiple [devices]({{<relref "devices.md">}}) can be added to the application.