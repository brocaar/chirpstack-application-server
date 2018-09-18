---
title: Requirements
menu:
    main:
        parent: install
        weight: 1
description: Instructions how to setup the LoRa App Server requirements.
---

# Requirements


## MQTT broker

LoRa App Server makes use of MQTT for publishing and receivng application
payloads. [Mosquitto](http://mosquitto.org/) is a popular open-source MQTT
server, but any MQTT broker implementing MQTT 3.1.1 should work.
In case you install Mosquitto, make sure you install a **recent** version.

### Install

#### Debian / Ubuntu

In order to install Mosquitto, execute the following command:

```bash
sudo apt-get install mosquitto
```

#### Other platforms

Please refer to the [Mosquitto download](https://mosquitto.org/download/) page
for information about how to setup Mosquitto for your platform.

## PostgreSQL database

LoRa App Server persists the gateway data into a
[PostgreSQL](https://www.postgresql.org) database. Note that PostgreSQL 9.5+
is required.

### pq_trgm extension

You also need to enable the [`pg_trgm`](https://www.postgresql.org/docs/current/static/pgtrgm.html)
(trigram) extension. Example to enable this extension (assuming your
LoRa App Server database is named `loraserver_as`):

Start the PostgreSQL prompt as the `postgres` user:

```bash
sudo -u postgres psql
```

Within the PostgreSQL prompt, enter the following queries:

```sql
-- change to the LoRa App Server database
\c loraserver_as

-- enable the extension
create extension pg_trgm;

-- exit the prompt
\q
```

### Install

#### Debian / Ubuntu

To install the latest PostgreSQL:

```bash
sudo apt-get install postgresql
```

#### Other platforms

Please refer to the [PostgreSQL download](https://www.postgresql.org/download/)
page for information how to setup PostgreSQL on your platform.

## Redis database

LoRa App Server stores all non-persistent data into a
[Redis](http://redis.io/) datastore. Note that at least Redis 2.6.0
is required.

### Install

#### Debian / Ubuntu

To Install Redis:

```bash
sudo apt-get install redis-server
```

#### Other platforms

Please refer to the [Redis](https://redis.io/) documentation for information
about how to setup Redis for your platform.
