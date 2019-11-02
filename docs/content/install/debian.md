---
title: Debian / Ubuntu
menu:
    main:
        parent: install
        weight: 3
description: Instructions how to install ChirpStack Application Server on a Debian or Ubuntu based Linux installation.
---

# Debian / Ubuntu installation

These steps have been tested with:

* Ubuntu 18.04 (LTS)
* Debian 10.0 (Buster)

### Creating an user and database

ChirpStack Application Server needs its **own** database. To create a new database,
start the PostgreSQL prompt as the `postgres` user:

{{<highlight bash>}}
sudo -u postgres psql
{{< /highlight >}}

Within the PostgreSQL prompt, enter the following queries:

{{<highlight sql>}}
-- create the chirpstack_as user
create role chirpstack_as with login password 'dbpassword';

-- create the chirpstack_as database
create database chirpstack_as with owner chirpstack_as;

-- enable the trigram and hstore extensions
\c chirpstack_as
create extension pg_trgm;
create extension hstore;

-- exit the prompt
\q
{{< /highlight >}}

To verify if the user and database have been setup correctly, try to connect
to it:

{{<highlight bash>}}
psql -h localhost -U chirpstack_as -W chirpstack_as
{{< /highlight >}}

## ChirpStack Network Server Debian repository

ChirpStack provides pre-compiled binaries packaged as Debian (.deb)
packages. In order to activate this repository, execute the following
commands:

{{<highlight bash>}}
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 1CE2AFD36DBCCA00

sudo echo "deb https://artifacts.chirpstack.io/packages/3.x/deb stable main" | sudo tee /etc/apt/sources.list.d/chirpstack.list
sudo apt-get update
{{< /highlight >}}

## Install ChirpStack Application Server

In order to install ChirpStack Application Server, execute the follwing command:

{{<highlight bash>}}
sudo apt-get install chirpstack-application-server
{{< /highlight >}}

After installation, modify the configuration file which is located at
`/etc/chirpstack-application-server/chirpstack-application-server.toml`. Given you used the password `dbpassword` when
creating the PostgreSQL database, you want to change the config variable
`postgresql.dsn` into:

`postgres://chirpstack_as:dbpassword@localhost/chirpstack_as?sslmode=disable`

An other required setting you must change is `application_server.external_api.jwt_secret`.

### Starting ChirpStack Application Server

How you need to (re)start and stop ChirpStack Application Server depends on if your
distribution uses init.d or systemd.

#### init.d

{{<highlight bash>}}
sudo /etc/init.d/chirpstack-application-server [start|stop|restart|status]
{{< /highlight >}}

#### systemd

{{<highlight bash>}}
sudo systemctl [start|stop|restart|status] chirpstack-application-server
{{< /highlight >}}

### ChirpStack Application Server log output

Now you've setup ChirpStack Application Server, it is a good time to verify that LoRa App
Server is actually up-and-running. This can be done by looking at the LoRa
App Server log output.

Like the previous step, the command you need to use for viewing the
log output depends on if your distribution uses init.d or systemd.

#### init.d

All logs are written to `/var/log/chirpstack-application-server/chirpstack-application-server.log`.
To view and follow this logfile:

{{<highlight bash>}}
tail -f /var/log/chirpstack-application-server/chirpstack-application-server.log
{{< /highlight >}}

#### systemd

{{<highlight bash>}}
journalctl -u chirpstack-application-server -f -n 50
{{< /highlight >}}

Example output:

{{<highlight text>}}
level=info msg="starting ChirpStack Application Server" docs="https://www.chirpstack.io/" version=3.3.0
level=info msg="storage: setting up storage package"
level=info msg="storage: setting up Redis pool"
level=info msg="storage: connecting to PostgreSQL database"
level=info msg="storage: applying PostgreSQL data migrations"
level=info msg="storage: PostgreSQL data migrations applied" count=0
level=info msg="integration/mqtt: TLS config is empty"
level=info msg="integration/mqtt: connecting to mqtt broker" server="tcp://localhost:1883"
level=info msg="integration/postgresql: connecting to PostgreSQL database"
level=info msg="integration/mqtt: connected to mqtt broker"
level=info msg="integration/mqtt: subscribing to tx topic" qos=0 topic=application/+/device/+/tx
level=info msg="api/as: starting application-server api" bind="0.0.0.0:8001" ca_cert= tls_cert= tls_key=
level=info msg="api/external: starting api server" bind="0.0.0.0:8080" tls-cert= tls-key=
level=info msg="api/external: registering rest api handler and documentation endpoint" path=/api
level=info msg="api/js: starting join-server api" bind="0.0.0.0:8003" ca_cert= tls_cert= tls_key=
level=info msg="metrics: starting prometheus metrics server" bind="127.0.0.1:7002"
{{< /highlight >}}

In case you see the following log messages, it means that ChirpStack Application Server
can't connect to [ChirpStack Network Server](https://www.chirpstack.io/network-server/).

{{<highlight text>}}
INFO[0001] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
INFO[0002] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
INFO[0005] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
{{< /highlight >}}

### Accessing ChirpStack Application Server

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
Run `chirpstack-application-server --help` for an overview of all available variables. Note
that configuration variables can be passed as cli arguments and / or environment
variables (which we did in the above example).

See [Configuration]({{< relref "config.md" >}}) for details on each config option.
