---
title: Debian / Ubuntu
menu:
    main:
        parent: install
        weight: 3
description: Instructions how to install LoRa App Server on a Debian or Ubuntu based Linux installation.
---

# Debian / Ubuntu installation

These steps have been tested with:

* Ubuntu 16.04 (LTS)
* Ubuntu 18.04 (LTS)
* Debian 9 (Stretch)

### Creating an user and database

LoRa App Server needs its **own** database. To create a new database,
start the PostgreSQL prompt as the `postgres` user:

{{<highlight bash>}}
sudo -u postgres psql
{{< /highlight >}}

Within the PostgreSQL prompt, enter the following queries:

{{<highlight sql>}}
-- create the loraserver_as user
create role loraserver_as with login password 'dbpassword';

-- create the loraserver_as database
create database loraserver_as with owner loraserver_as;

-- enable the trigram extension
\c loraserver_as
create extension pg_trgm;

-- exit the prompt
\q
{{< /highlight >}}

To verify if the user and database have been setup correctly, try to connect
to it:

{{<highlight bash>}}
psql -h localhost -U loraserver_as -W loraserver_as
{{< /highlight >}}

## LoRa Server Debian repository

The LoRa Server project provides pre-compiled binaries packaged as Debian (.deb)
packages. In order to activate this repository, execute the following
commands:

{{<highlight bash>}}
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 1CE2AFD36DBCCA00

sudo echo "deb https://artifacts.loraserver.io/packages/3.x/deb stable main" | sudo tee /etc/apt/sources.list.d/loraserver.list
sudo apt-get update
{{< /highlight >}}

## Install LoRa App Server

In order to install LoRa App Server, execute the follwing command:

{{<highlight bash>}}
sudo apt-get install lora-app-server
{{< /highlight >}}

After installation, modify the configuration file which is located at
`/etc/lora-app-server/lora-app-server.toml`. Given you used the password `dbpassword` when
creating the PostgreSQL database, you want to change the config variable
`postgresql.dsn` into:

`postgres://loraserver_as:dbpassword@localhost/loraserver_as?sslmode=disable`

An other required setting you must change is `application_server.external_api.jwt_secret`.

### Starting LoRa App Server

How you need to (re)start and stop LoRa App Server depends on if your
distribution uses init.d or systemd.

#### init.d

{{<highlight bash>}}
sudo /etc/init.d/lora-app-server [start|stop|restart|status]
{{< /highlight >}}

#### systemd

{{<highlight bash>}}
sudo systemctl [start|stop|restart|status] lora-app-server
{{< /highlight >}}

### LoRa App Server log output

Now you've setup LoRa App Server, it is a good time to verify that LoRa App
Server is actually up-and-running. This can be done by looking at the LoRa
App Server log output.

Like the previous step, the command you need to use for viewing the
log output depends on if your distribution uses init.d or systemd.

#### init.d

All logs are written to `/var/log/lora-app-server/lora-app-server.log`.
To view and follow this logfile:

{{<highlight bash>}}
tail -f /var/log/lora-app-server/lora-app-server.log
{{< /highlight >}}

#### systemd

{{<highlight bash>}}
journalctl -u lora-app-server -f -n 50
{{< /highlight >}}

Example output:

{{<highlight text>}}
Sep 25 12:44:59 ubuntu-xenial systemd[1]: Started lora-app-server.
level=info msg="starting LoRa App Server" docs="https://docs.loraserver.io/" version=0c5ba3f
level=info msg="connecting to postgresql"
level=info msg="setup redis connection pool"
level=info msg="handler/mqtt: connecting to mqtt broker" server="tcp://localhost:1883"
level=info msg="connecting to network-server api" ca-cert= server="127.0.0.1:8000" tls-cert= tls-key=
level=info msg="handler/mqtt: connected to mqtt broker"
level=info msg="handler/mqtt: subscribling to tx topic" topic="application/+/node/+/tx"
level=info msg="applying database migrations"
level=info msg="migrations applied" count=0
level=info msg="migrating node-session data from Redis"
level=info msg="starting application-server api" bind="127.0.0.1:8001" ca-cert= tls-cert= tls-key=
level=warning msg="client api authentication and authorization is disabled (set jwt-secret to enable)"
level=info msg="starting client api server" bind="0.0.0.0:8080" tls-cert="/opt/lora-app-server/certs/http.pem" tls-key="/opt/lora-app-server/certs/http-key.pem"
level=info msg="registering rest api handler and documentation endpoint" path="/api"
{{< /highlight >}}

In case you see the following log messages, it means that LoRa App Server
can't connect to [LoRa Server](https://docs.loraserver.io/loraserver/).

{{<highlight text>}}
INFO[0001] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
INFO[0002] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
INFO[0005] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
{{< /highlight >}}

### Accessing LoRa App Server

If a TLS certificate has been configured (optional), use http**s://**
else use the http:// option (default).

#### HTTP

* **Web-interface** [http://localhost:8080/](http://localhost:8080/)
* **API** [http://localhost:8080/api](http://localhost:8080/api)

#### HTTPS

* **Web-interface** [https://localhost:8080/](https://localhost:8080/)
* **API** [https://localhost:8080/api](https://localhost:8080/api)

## Configuration

In the example above, we've just touched a few configuration variables.
Run `lora-app-server --help` for an overview of all available variables. Note
that configuration variables can be passed as cli arguments and / or environment
variables (which we did in the above example).

See [Configuration]({{< relref "config.md" >}}) for details on each config option.
