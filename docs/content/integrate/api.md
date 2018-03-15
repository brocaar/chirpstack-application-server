---
title: API
menu:
    main:
        parent: integrate
        weight: 3
---

# LoRa App Server API

LoRa App Server can be easily integrated into your own project using
either the gRPC or the JSON REST API. The API has been divided into
the following services / endpoints:

### Application service

For application management and the management of users assigned to an
application.

### DownlinkQueue service

For management of the downlink queue / sending downlink data to nodes.

### Gateway service

For management of gateways and retrieving gateway statistics.

### Internal service

For "internal" LoRa App Server specific actions (e.g. login and retrieving
the user profile). These endpoints should not be used for integration and
are there to facilidate LoRa App Server specific tasks.

### Node service

For management and (ABP) activation of nodes / reading node activation details.

### Organization service

For management of organizations and the management of users assigned to an
organization.

### User service

For management of users.
