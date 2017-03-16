# Getting started

A complete LoRa Server setup, requires the setup of the following components:

* [LoRa Gateway Bridge](https://docs.loraserver.io/lora-gateway-bridge/)
* [LoRa Server](https://docs.loraserver.io/loraserver/)
* [LoRa App Server](https://docs.loraserver.io/lora-app-server/)

This getting started document describes the steps needed to setup LoRa App
Server using the provided Debian package repository. Please note that LoRa
App Server is not limited to Debian / Ubuntu only! General purpose binaries
can be downloaded from the 
[releases](https://github.com/brocaar/lora-app-server/releases) page.

!!! info
	An alternative way to setup all the components is by using the
	[loraserver-setup](https://github.com/brocaar/loraserver-setup) Ansible
	scripts. It automates the steps below and can also be used in combination
	with [Vagrant](https://www.vagrantup.com/).

!!! warning
    This getting started guide does not cover setting up firewall rules! After
    setting up LoRa App Server and its requirements, don't forget to configure
    your firewall rules.

## Setting up LoRa App Server

These steps have been tested with:

* Debian Jessie
* Ubuntu Trusty (14.04)
* Ubuntu Xenial (16.06)

### LoRa Server Debian repository

The LoRa Server project provides pre-compiled binaries packaged as Debian (.deb)
packages. In order to activate this repository, execute the following
commands:

```bash
source /etc/lsb-release
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 1CE2AFD36DBCCA00
sudo echo "deb https://repos.loraserver.io/${DISTRIB_ID,,} ${DISTRIB_CODENAME} testing" | sudo tee /etc/apt/sources.list.d/loraserver.list
sudo apt-get update
```

### MQTT broker

LoRa App Server makes use of MQTT for publishing and receivng application
payloads. [Mosquitto](http://mosquitto.org/) is a popular open-source MQTT
server. Make sure you install a **recent** version of Mosquitto.

For Ubuntu Trusty (14.04), execute the following command in order to add the
Mosquitto Apt repository:

```bash
sudo apt-add-repository ppa:mosquitto-dev/mosquitto-ppa
sudo apt-get update
```

In order to install Mosquitto, execute the following command:

```bash
sudo apt-get install mosquitto
```

### Redis

LoRa App Server stores all non-persistent data into a
[Redis](http://redis.io/) datastore. Note that at least Redis 2.6.0
is required. To Install Redis:

```bash
sudo apt-get install redis-server
```

### PostgreSQL server

LoRa App Server stores all persistent data into a
[PostgreSQL](http://www.postgresql.org/) database. To install PostgreSQL:

```bash
sudo apt-get install postgresql
```

#### Creating a loraserver user and database

Start the PostgreSQL promt as the `postgres` user:

```bash
sudo -u postgres psql
```

Within the the PostgreSQL promt, enter the following queries:

```sql
-- create the loraserver user with password "dbpassword"
create role loraserver with login password 'dbpassword';

-- create the loraserver database
create database loraserver with owner loraserver;

-- exit the prompt
\q
```

To verify if the user and database have been setup correctly, try to connect
to it:

```bash
psql -h localhost -U loraserver -W loraserver
```

### Install LoRa App Server

In order to install LoRa App Server, execute the follwing command:

```bash
sudo apt-get install lora-app-server
```

After installation, modify the configuration file which is located at
`/etc/default/lora-app-server`.

Given you used the password `dbpassword` when creating the PostgreSQL database,
you want to change the config variable `POSTGRES_DSN` into:

```
POSTGRES_DSN=postgres://loraserver:dbpassword@localhost/loraserver?sslmode=disable
```

An other required setting you must change is `JWT_SECRET`.

### Starting LoRa App Server

How you need to (re)start and stop LoRa Gateway Bridge depends on if your
distribution uses init.d or systemd.

#### init.d

```bash
sudo /etc/init.d/lora-app-server [start|stop|restart|status]
```

#### systemd

```bash
sudo systemctl [start|stop|restart|status] lora-app-server
```

### LoRa App Server log output

Now you've setup LoRa App Server, it is a good time to verify that LoRa App
Server is actually up-and-running. This can be done by looking at the LoRa
App Server log output.

Like the previous step, the command you need to use for viewing the
log output depends on if your distribution uses init.d or systemd.

#### init.d

All logs are written to `/var/log/lora-app-server/lora-app-server.log`.
To view and follow this logfile:

```bash
tail -f /var/log/lora-app-server/lora-app-server.log
```

#### systemd

```bash
journalctl -u lora-app-server -f -n 50
```

Example output:

```
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
```

In case you see the following log messages, it means that LoRa App Server
can't connect to [LoRa Server](https://docs.loraserver.io/loraserver/).

```
INFO[0001] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
INFO[0002] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
INFO[0005] grpc: addrConn.resetTransport failed to create client transport: connection error: desc = "transport: dial tcp 127.0.0.1:8000: getsockopt: connection refused"; Reconnecting to {"127.0.0.1:8000" <nil>}
```

### Accessing LoRa App Server

To access the web-interface, point your browser to
[https://localhost:8080](https://localhost:8080). Note that it is normal that
this will raise a security warning, as a self-signed certificate is being used.
To login, use *admin* / *admin* (don't forget to reset this password!).

To access the REST API endpoint, point your browser to
[https://localhost:8080/api](https://localhost:8080/api).

## Configuration

In the example above, we've just touched a few configuration variables.
Run `lora-app-server --help` for an overview of all available variables. Note
that configuration variables can be passed as cli arguments and / or environment
variables (which we did in the above example).

See [Configuration](configuration.md) for details on each config option.
