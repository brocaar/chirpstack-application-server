---
title: HTTP
menu:
    main:
        parent: sending-receiving
---

# HTTP integration

When configured, the HTTP integration will make `POST` requests to the
configured event endpoint or endpoints (multiple URLs can ben configured, comma
separated). The `eventType` URL query parameter indicates the type of the event.

## Marshaler

It is possible to select from multiple marshalers:

* JSON: This uses the [Protocol Buffers JSON mapping](https://developers.google.com/protocol-buffers/docs/proto3#json)
* Protocol Buffers: This uses the binary [Protocol Buffers](https://developers.google.com/protocol-buffers) encoding
* JSON v3: This uses the legacy JSON mapping. This option will be removed in the next major release.

## Events

The HTTP integration exposes all events as documented by [Event Types](../#event-types).
