---
title: Authentication
menu:
    main:
        parent: integrate
        weight: 4
description: Information on API tokens for authentication and authorization.
---

# Authentication / authorization

Both the gRPC and the JSON REST interface are protected by an authentication
and authorization meganism. For this [JSON web-tokens](https://jwt.io) are
used, using the `--jwt-secret` / `JWT_SECRET` value for signing. Therefore
it is important to choose an unique and strong secret.

To generate a random secret, you could use the following command:

{{<highlight bash>}}
openssl rand -base64 32
{{< /highlight >}}

Example claim:

`HEADER:ALGORITHM & TOKEN TYPE`
{{<highlight json>}}
{
	"alg": "HS256",
	"typ": "JWT"
}
{{< /highlight >}}

`PAYLOAD:DATA`
{{<highlight json>}}
{
	"iss": "lora-app-server",      // issuer of the claim
	"aud": "lora-app-server",      // audience for which the claim is intended
	"nbf": 1489566958,             // unix time from which the token is valid
	"exp": 1489653358,             // unix time when the token expires
	"sub": "user",                 // subject of the claim (an user)
	"username": "admin"            // username the client claims to be
}
{{< /highlight >}}

`VERIFY SIGNATURE`
{{<highlight html>}}
HMACSHA256(
	base64UrlEncode(header) + "." +
	base64UrlEncode(payload),
	your-256-bit-secret
) []secret base64 encoded
{{< /highlight >}}

**Notes:**

* Update `nbf` and `exp`
* Enter `your-256-bit-secret`
* Uncheck `secret base64 encoded`

## Setting the authentication token

### gRPC

When using [gRPC](http://grpc.io/), the JWT token needs to be stored in the
`authorization` key of the request metadata. For example in Go, this can be
done by the [grpc.WithPerRPCCredentials](https://godoc.org/google.golang.org/grpc#WithPerRPCCredentials)
method.

### REST API

For requests to the RESTful JSON interface, you need to set the JWT token
using the `Grpc-Metadata-Authorization` header field. The token needs to
be present for each request.
