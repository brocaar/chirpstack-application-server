---
title: gRPC
menu:
    main:
        parent: integrate
        weight: 5
description: Information about integrating with the gRPC API interface.
---

# gRPC API

ChirpStack Application Server provides a [gRPC](http://www.grpc.io/) API for easy integration
with your own projects. gRPC is a RPC framework on top of [protocol-buffers](https://developers.google.com/protocol-buffers/).
gRPC is really easy to work with, as the protocol buffer file can be seen as
a contract between the provider and consumer, in other words the fields and
their datatypes are known.

The gRPC server is listening on the port configured in the
`application_server.external_api.bind` configuration.

Using the gRPC toolset, it is possible to generate client code for the following
languages (officially suported by gRPC):

* C++
* Go (included)
* Node.js
* Java
* Ruby
* Android Java
* PHP
* Python
* C#
* Objective-C

## Links

* [gRPC documentation](http://www.grpc.io/)
* [ChirpStack Application Server .proto files](https://github.com/brocaar/chirpstack-api/tree/master/protobuf/as/external/api)
* [ChirpStack Application Server Go client](https://godoc.org/github.com/brocaar/chirpstack-api/go/as/external/api)
