---
title: Organizations
menu:
    main:
        parent: use
        weight: 4
description: Manage Organizations, Service Profiles assigned to Organizations and Organization Users.
---

# Organization management

An organization can be used to let organizations or teams manage their
own applications and optionally their own gateways.

An organization can have:

* Service Profiles
* Device Profiles
* Gateways (when allowed)
* Applications
* Users

## Service Profiles

Global Admin Users are able to manage
[Service Profiles]({{<relref "service-profiles.md">}}) for any 
organization. These Service Profiles can then be used by (organization)
admin users when creating applications.

## Device Profiles

[Device Profiles]({{<relref "device-profiles.md">}}) can be created by
(organization) admin users and can be assigned when creating a
[Device]({{<relref "devices.md">}}).

## Gateways

An organization can manage its own set of gateways. Note that when an organization
is created by a global administrator, it can decide that an organization can not
have any gateways. In this case, the gateway option is not available to an
organization.

That an organization is able to manage its own set of gateways does not mean
that the coverage is limited to this set of gateways. Gateways connectivity
will be shared across the whole network.

## Applications

[Applications]({{<relref "applications.md">}}) can be created by (organization)
admin users and define a group of devices with the same purpose.

## Users

Users can be assigned to an organization to grant them access to the
organization. Within the context of that assigment, an user can be an
organization administrator or a regular user.

### Organization administrator

An organization administrator is authorized to manage the users assigned
with the organization and manage the gateways, applications and nodes of the
gateway.

### Regular user

Regular users are able to see all data, but are not able to make any
modifications.
