---
title: HTTP
menu:
    main:
        parent: sending-receiving
---

# HTTP integration

When configured, the HTTP integration will make `POST` requests
to the configured endpoints on the following events:

* Received uplink data
* Status notifications
* Join notifications
* ACK notifications
* Error notifications

## Events

The HTTP integration exposes all events as documented by [Event Types](../#event-types).
