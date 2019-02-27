---
title: AWS Simple Notification Service (SNS)
menu:
  main:
    parent: sending-receiving
---

# AWS Simple Notification Service (SNS)

The [Simple Notification Service (SNS)](https://aws.amazon.com/sns/) integration
publishes all the events to a SNS Topic to which other applications or AWS
services can subscribe for further processing.

## Events

The AWS Simple Notification Service integration exposes all events as
documented by [Event types](../#event-types).

## Message attributes

The following message attributes are added to each published message:

* `event` - the event type
* `dev_eui` - the device EUI
* `application_id` - the LoRa App Server application ID

