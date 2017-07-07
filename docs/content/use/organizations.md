---
title: Organizations
menu:
    main:
        parent: use
        weight: 3
---

## Organization management

An organization can be used to let organizations or teams manage their
own applications and optionally their own gateways.

An organization has:

* Users
* Gateways (when allowed)
* Applications

### Users

Users can be assigned to an organization to grant them access to the
organization. Within the context of that assigment, an user can be an 
organization administrator or a regular user.


#### Organization administrator

An organization administrator is authorized to manage the users assigned
with the organization and manage the gateways, applications and nodes of the
gateway.

#### Regular user

Regular users are able to see all data, but are not able to make any
modifications.

### Gateways

An organization can manage its own set of gateways. Note that when an organization
is created by a global administrator, it can decide that an organization can not
have any gateways. In this case, the gateway option is not available to an
organization.

That an organization is able to manage its own set of gateways does not mean
that the coverage is limited to this set of gateways. Gateways connectivity
will be shared across the whole network.
