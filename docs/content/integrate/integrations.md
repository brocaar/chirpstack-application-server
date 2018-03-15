---
title: Data integrations
menu:
    main:
        parent: integrate
        weight: 2
---

# Data integrations

## HTTP

LoRa App Server provides the possibility to configure HTTP integrations
per application. When in the application screen, click on the *Integrations*
tab and then click *Add integration*. Then click the *HTTP Integration*
option.

The HTTP integration follows exaclty the same JSON data structure as the
data structures documented in the [Send / receive data]({{< ref "data.md" >}})
documentation.

The following endpoints can be configured:

* Uplink data
* Join notifications
* ACK notifications
* Error notifications

LoRa App Server will use the `POST` HTTP method.