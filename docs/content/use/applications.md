---
title: Applications
menu:
    main:
        parent: use
        weight: 4
---

## Application management

An application is a collection of nodes with the same purpose / of the same type.
Think of a weather station collecting data at different locations for example.

An application has:

* Users
* Nodes

### Users

Users can be assigned to an application to grant them access to the
application. Note that when an user is already member of the organization
that these permissions are inherited automatically. In that case it is not
needed to also add the user to the application. Within the context of an
application two types of users exists.

#### Application administrator

Application administrators is authorized to manage the users assigned to the
application and to manage the nodes.

#### Regular user

Regular users are a ble to see all data within the application, but are not
able to make any modifications.

### Network settings

An application can hold the network settings for all nodes within the
application. This makes it easy to keep the configuration of all nodes
in sync. These settings are identical to the settings on the node. For all
available options, refer to the [nodes]({{< relref "nodes.md" >}}) documentation.
