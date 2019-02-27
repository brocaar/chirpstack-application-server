---
title: Azure Service Bus
menu:
  main:
    parent: sending-receiving
---

# Azure Service Bus

The [Azure Service Bus](https://azure.microsoft.com/en-us/services/service-bus/)
integration publishes all the events to a Service Bus [Topic or Queue](https://docs.microsoft.com/en-us/azure/service-bus-messaging/service-bus-messaging-overview)
to which applications can subscribe.

## Events

The Azure Service Bus integration exposes all events as documented by [Event types](../#event-types).

## User properties

The following user properties are added to each published message:

* `event` - the event type
* `dev_eui` - the device EUI
* `application_id` - the LoRa App Server application ID

