---
title: RESTful JSON
menu:
    main:
        parent: integrate
        weight: 6
description: Information about integrating with the RESTful JSON API.
---

# RESTful JSON API

Besides the [gRPC]({{< relref "grpc.md" >}}) API, LoRa App Server also provides
a JSON REST API. As this is technically a "gateway" between the gRPC API,
it contains exactly the same calls (using RESTful resources) and exposes the
same fields.

## API console

LoRa App Server comes with an API console (based on Swagger UI) containing all
API endpoints and their documentation. This console is accessible at `/api`.
Please note that for most API endpoints, you need to provide a valid JWT token
which can be entered in the right top input field. See also
[authentication]({{< relref "auth.md" >}}).

![Swagger API](/lora-app-server/img/swagger.png)
