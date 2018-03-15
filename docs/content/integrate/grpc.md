---
title: gRPC
menu:
    main:
        parent: integrate
        weight: 5
---

# gRPC API

LoRa App Server provices a [gRPC](http://www.grpc.io/) API for easy integration
with your own projects. gRPC is a RPC framework on top of [protocol-buffers](https://developers.google.com/protocol-buffers/).
gRPC is really easy to work with, as the protocol buffer file can be seen as
a contract between the provider and consumer, in other words the fields and
their datatypes are known.

The gRPC server is listening on the port configured in the
`--http-bind` / `HTTP_BIND` configuration.

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
* [LoRa App Server .proto files](https://github.com/brocaar/lora-app-server/tree/master/api)
* [LoRa App Server Go client](https://godoc.org/github.com/brocaar/lora-app-server/api)

## Code examples

### Go

```go
package main

import (
	"context"
	"log"

	"github.com/brocaar/lora-app-server/api"
	"google.golang.org/grpc"
)

func main() {
	// allow insecure / non-tls connections
	grpcOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	// connect to lora-app-server
	asConn, err := grpc.Dial("localhost:8080", grpcOpts...)
	if err != nil {
		log.Fatal(err)
	}

	// create a new node
	nodeAPI := api.NewNodeClient(asConn)
	_, err = nodeAPI.Create(context.Background(), &api.CreateNodeRequest{
		DevEUI:                 "0101010101010101",
		AppEUI:                 "0202020202020202",
		AppKey:                 "03030303030303030303030303030303",
		ApplicationID:          1,
		UseApplicationSettings: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}
```
