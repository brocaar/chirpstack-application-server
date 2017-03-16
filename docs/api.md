# API

LoRa App Server provides a [gRPC](http://www.grpc.io/) API for easy integration
with your own services. On top of this gRPC API, LoRa App Server provides a
RESTful JSON interface, so that you can use the API for web-applications
(note that gRPC is a binary API layer, on top of
[protocol-buffers](https://developers.google.com/protocol-buffers/).

!!! info "Protocol-buffer files"
    LoRa App Server provides a gRPC client for Go. For other programming languages
    you can use the .proto files inside the [api](https://github.com/brocaar/lora-app-server/tree/master/api)
    folder for generating clients. See the [gRPC](http://www.grpc.io/) documentation
    for documentation.

## RESTful JSON interface

Since gRPC [can't be used in browsers yet](http://www.grpc.io/faq/), LoRa App
Server provides a RESTful JSON API (by using [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway))
on top of the gRPC API, exposing the same API methods as the gRPC API.
The REST API documentation and interactive console can be found at `/api`.

![Swagger API](img/swagger.png)

## Authentication and authorization

Both the gRPC and RESTful JSON interface are protected by an authentication
and authorization meganism. For this [JSON web-tokens](https://jwt.io) are
used, using the `--jwt-secret` / `JWT_SECRET` value for signing. Therefore
it is important to choose an unique and strong secret.

To generate a random secret, you could use the following command:

```bash
openssl rand -base64 32
```

Example claim:

```json
{
	"iss": "lora-app-server",      // issuer of the claim
	"aud": "lora-app-server",      // audience for which the claim is intended
	"nbf": 1489566958,             // unix time from which the token is valid
	"exp": 1489653358,             // unix time when the token expires
	"sub": "user",                 // subject of the claim (an user)
	"username": "admin"            // username the client claims to be
}
```

### Users

!!! warning
	An initial user named *admin* with password *admin* will be created when
	installing LoRa App Server. Make sure to change this password immediately!

Users can be created using either the API or web-interface. When creating
an admin user, the user will be able to create other (admin) users,
applications and assign users to applications. Global admin users have
system-wide access. Regular users will only gain access when assigned to
applications.

### Application users

To give users access to specific applications, an user can be assigned to
one or multiple applications (again either using the API or web-interface).

An admin user (within the context of an application) is able to:

- assign or create other (admin, within the context of the application) users to that application
- add, delete and modify nodes

A regular users of an application is able to:

- view all the nodes

### Setting the authentication token

For requests to the RESTful JSON interface, you need to set the JWT token
using the `Grpc-Metadata-Authorization` header field. The token needs to
be present for each request.

When using [gRPC](http://grpc.io/), the JWT token needs to be stored in the
`authorization` key of the request metadata. For example in Go, this can be
done by the [grpc.WithPerRPCCredentials](https://godoc.org/google.golang.org/grpc#WithPerRPCCredentials)
method.

## Security / TLS

The http server for serving the web-interface and API (both gRPC as the
RESTful JSON api) must be secured by using a TLS certificate.

### Self-signed certificate

A self-signed certificate can be generated with the following command:

```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 90 -nodes
```

### Let's Encrypt

For generating a certificate with [Let's Encrypt](https://letsencrypt.org/),
first follow the [getting started](https://letsencrypt.org/getting-started/)
instructions. When the `letsencrypt` cli tool has been installed, execute:

```bash
letsencrypt certonly --standalone -d DOMAINNAME.HERE 
```
