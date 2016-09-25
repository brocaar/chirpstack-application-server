# Getting started

This getting started document describes the steps needed to setup LoRa App
Server and all its requirements on Ubuntu 16.04 LTS. When using an other Linux
distribution, you might need to adapt these steps.

!!! warning
    This getting started guide does not cover setting up firewall rules! After
    setting up LoRa Server and its requirements, don't forget to configure
    your firewall rules.

## MQTT broker

LoRa App Server makes use of MQTT for publishing and receiving application
payloads. [Mosquitto](http://mosquitto.org/) is a
popular open-source MQTT broker. Make sure you install a recent version of
Mosquitto (the Mosquitto project provides repositories for various Linux
distributions). Ubuntu 16.04 LTS already includes a recent version which can be
installed with:

```bash
sudo apt-get install mosquitto
```

## Redis

LoRa Server stores all session-related and non-persistent data into a
[Redis](http://redis.io/) datastore. Note that at least Redis 2.6.0 is required.
To Install Redis:

```bash
sudo apt-get install redis-server
```

## PostgreSQL server

LoRa App Server stores all persistent data into a
[PostgreSQL](http://www.postgresql.org/) database. To install PostgreSQL:

```bash
sudo apt-get install postgresql
```

### Creating a loraserver user and database

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

## Install LoRa App Server

Create a system user for LoRa App Server:

```bash
sudo useradd -M -r -s /bin/false appserver
```

Download and unpack a pre-compiled binary from the
[releases](https://github.com/brocaar/lora-app-server/releases) page:

```bash
# replace VERSION with the latest version or the version you want to install

# download
wget https://github.com/brocaar/lora-app-server/releases/download/VERSION/lora_app_server_VERSION_linux_amd64.tar.gz

# unpack
tar zxf lora_app_server_VERSION_linux_amd64.tar.gz

# move the binary to /opt/lora-app-server/bin
sudo mkdir -p /opt/lora-app-server/bin
sudo mv lora-app-server /opt/lora-app-server/bin
```

As the web-interface and API requires a TLS certificate, create a self-signed
certificate:

```bash
# create cert directory
sudo mkdir -p /opt/lora-app-server/certs

# generate self-signed certificate
sudo openssl req -x509 -newkey rsa:4096 -keyout /opt/lora-app-server/certs/http-key.pem -out /opt/lora-app-server/certs/http.pem -days 365 -nodes
```

In order to start LoRa App Server as a service, create the file
`/etc/systemd/system/lora-app-server.service` with as content:

```
[Unit]
Description=lora-app-server
After=mosquitto.service postgresql.service

[Service]
User=appserver
Group=appserver
ExecStart=/opt/lora-app-server/bin/lora-app-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

In order to configure LoRa App Server, create a directory named
`/etc/systemd/system/lora-app-server.service.d`:

```bash
sudo mkdir /etc/systemd/system/lora-app-server.service.d
```

Inside this directory, put a file named `lora-app-server.conf`:

```
[Service]
Environment="BIND=127.0.0.1:8001"
Environment="HTTP_BIND=0.0.0.0:8080"
Environment="NS_SERVER=127.0.0.1:8000"
Environment="HTTP_TLS_CERT=/opt/lora-app-server/certs/http.pem"
Environment="HTTP_TLS_KEY=/opt/lora-app-server/certs/http-key.pem"
Environment="DB_AUTOMIGRATE=True"
Environment="MIGRATE_NODE_SESSIONS=True"
Environment="REDIS_URL=redis://localhost:6379"
Environment="POSTGRES_DSN=postgres://loraserver:dbpassword@localhost/loraserver?sslmode=disable"
```

## Starting LoRa App Server

```bash
# start
sudo systemctl start lora-app-server

# restart
sudo systemctl restart lora-app-server

# stop
sudo systemctl stop lora-app-server
```

Verifiy that LoRa App Server is up-and running by looking at its log-output:

```bash
journalctl -u lora-app-server -f -n 50
```

The log should be something like:

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

To access the webinterface, point your browser to
[https://localhost:8080](https://localhost:8080). Note that it is normal that
this will raise a security warning, as a self-signed certificate is being used.

To access the REST API endpoint, point your browser to
[https://localhost:8080/api](https://localhost:8080/api).

## Configuration

In the example above, we've just touched a few configuration variables.
Run `lora-app-server --help` for an overview of all available variables. Note
that configuration variables can be passed as cli arguments and / or environment
variables (which we did in the above example).

See [Configuration](configuration.md) for details on each config option.
