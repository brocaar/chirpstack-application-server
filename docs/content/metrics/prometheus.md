---
title: Prometheus
menu:
  main:
    parent: metrics
    weight: 1
description: Read metrics from the Prometheus metrics endpoint.
---

# Prometheus metrics

ChirpStack Application Server provides a [Prometheus](https://prometheus.io/) metrics endpoint
for monitoring the performance of the ChirpStack Application Server service. Please refer to
the [Prometheus](https://prometheus.io/) website for more information on
setting up and using Prometheus.

## Configuration

Please refer to the [Configuration documentation]({{<ref "install/config.md">}}).

## Metrics

### Go runtime metrics

These metrics are prefixed with `go_` and provide general information about
the process like:

* Garbage-collector statistics
* Memory usage
* Go go-routines

### gRPC API metrics

These metrics are prefixed with `grpc_` and provide metrics about the gRPC
API (both the external gRPC / REST API and the gRPC API used by [ChirpStack Network Server](/network-server/)), e.g.:

* The number of times each API was called
* The duration of each API call (if enabled in the [Configuration]({{<ref "install/config.md">}}))


### Join-Server API

These metrics are prefixed with `api_joinserver_` and provide metrics about
the ChirpStack Application Server Join-Server API endpoint. This endpoint is used by
ChirpStack Network Server when an OTAA Join-Request is received.

* The number of times each API was called
* The duration of each API call (if enabled in the [Configuration]({{<ref "install/config.md">}}))
