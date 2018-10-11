---
title: Google Cloud Platform Pub/Sub
menu:
    main:
        parent: sending-receiving
---

# Google Cloud Platform Pub/Sub

The [Google Cloud Platform](https://cloud.google.com/) [Pub/Sub](https://cloud.google.com/pubsub/)
integration publishes all the events to a configurable GCP Pub/Sub topic.
Using the GCP console (or APIs) you are able to create one or multiple Pub/Sub
subscriptions, for integrating this with your application(s) or store the data
in one of the storage options provided by the Google Cloud Platform.

## Events

The GCP Pub/Sub integration exposes all events as documented by [Event Types](../#event-types).

## Attributes

The following attributes are added to each Pub/Sub message:

* `event`: the event type
* `devEUI`: the device EUI to which the event relates
