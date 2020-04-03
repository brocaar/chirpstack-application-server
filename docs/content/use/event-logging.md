---
title: Event logging
menu:
    main:
        parent: use
        weight: 12
description: Show live device events for debugging device behavior.
---

# Event logging

ChirpStack Application Server makes it possible to log events sent to the MQTT broker
or configured integrations. To use this feature, you fist need to go to
the device detail page. Once you are on this page, open the **Device Data**
tab.

**Note:** This is for debugging purposes only! Do not use this for integration
with your applications.

As soon as you open this page, ChirpStack Application Server will subscribe to the events
of the selected device. Once an event is received, it will be displayed
without the need to refresh the page.

## Exposed events

Note that all the displayed data can be expanded by clicking on each key.
E.g. **> payload: {} 9 keys** means you can expand this **payload**
item as it has nine sub-items.

The payloads that are exposed are documented by the
[Sending and Receiving]({{<ref "integrate/sending-receiving/mqtt.md">}}) page.
You will also find examples on this page.
