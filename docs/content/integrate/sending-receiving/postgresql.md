---
title: PostgreSQL
menu:
  main:
    parent: sending-receiving
---

# PostgreSQL

The [PostgreSQL](https://www.postgresql.org/) integration writes all events
into a PostgreSQL database. This database can then be used by other
applications or visualized using for example [Grafana](https://grafana.com/)
using the [PostgreSQL Data Source](https://grafana.com/docs/features/datasources/postgres/#using-postgresql-in-grafana).

* ChirpStack Application Server will not create these tables for you. Create statements are
  given below.
* You must enable the `hstore` extension for this database table, this can be
  done with the SQL statement: `create extension hstore;`.
* This database does not have to be the same database as used by
  ChirpStack Application Server.
* This may generate a lot of data depending the number of devices and number
  of messages sent per device.

## Create database example

Please see below an example for creating the PostgreSQL database. Depending
your PostgreSQL installation, these commands might be different.

Enter the PostgreSQL as the `postgres` user:

{{<highlight bash>}}
sudo -u postgres psql
{{< /highlight >}}

Within the PostgreSQL prompt, enter the following queries:

{{<highlight sql>}}
-- create the chirpstack_as_events user
create role chirpstack_as_events with login password 'dbpassword';

-- create the chirpstack_as_events database
create database chirpstack_as_events with owner chirpstack_as_events;

-- enable the hstore extension
\c chirpstack_as_events
create extension hstore;

-- exit the prompt
\q
{{< /highlight >}}

To verify if the user and database have been setup correctly, try to connect
to it:

{{<highlight bash>}}
psql -h localhost -U chirpstack_as_events -W chirpstack_as_events
{{< /highlight >}}

## Events

### Uplink data

Uplink data is written into the table `device_up`. The following schema
must exist:

{{<highlight sql>}}
create table device_up (
	id uuid primary key,
	received_at timestamp with time zone not null,
	dev_eui bytea not null,
	device_name varchar(100) not null,
	application_id bigint not null,
	application_name varchar(100) not null,
	frequency bigint not null,
	dr smallint not null,
	adr boolean not null,
	f_cnt bigint not null,
	f_port smallint not null,
	tags hstore not null,
	data bytea not null,
	rx_info jsonb not null,
	object jsonb not null
);

-- NOTE: These are example indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_up_received_at on device_up(received_at);
create index idx_device_up_dev_eui on device_up(dev_eui);
create index idx_device_up_application_id on device_up(application_id);
create index idx_device_up_frequency on device_up(frequency);
create index idx_device_up_dr on device_up(dr);
create index idx_device_up_tags on device_up(tags);
{{</highlight>}}

### Device status

Device-status data is written into the table `device_status`. The following
schema must exist:

{{<highlight sql>}}
create table device_status (
	id uuid primary key,
	received_at timestamp with time zone not null,
	dev_eui bytea not null,
	device_name varchar(100) not null,
	application_id bigint not null,
	application_name varchar(100) not null,
	margin smallint not null,
	external_power_source boolean not null,
	battery_level_unavailable boolean not null,
	battery_level numeric(5, 2) not null,
	tags hstore not null
);

-- NOTE: These are example indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_status_received_at on device_status(received_at);
create index idx_device_status_dev_eui on device_status(dev_eui);
create index idx_device_status_application_id on device_status(application_id);
create index idx_device_status_tags on device_status(tags);
{{</highlight>}}

### Join

Join notifications are written into the table `device_join`. The following
schema must exist:

{{<highlight sql>}}
create table device_join (
	id uuid primary key,
	received_at timestamp with time zone not null,
	dev_eui bytea not null,
	device_name varchar(100) not null,
	application_id bigint not null,
	application_name varchar(100) not null,
	dev_addr bytea not null,
	tags hstore not null
);

-- NOTE: These are example indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_join_received_at on device_join(received_at);
create index idx_device_join_dev_eui on device_join(dev_eui);
create index idx_device_join_application_id on device_join(application_id);
create index idx_device_join_tags on device_join(tags);
{{</highlight>}}

### ACK

ACK notifications are written into the table `device_ack`. The following schema
must exist:

{{<highlight sql>}}
create table device_ack (
	id uuid primary key,
	received_at timestamp with time zone not null,
	dev_eui bytea not null,
	device_name varchar(100) not null,
	application_id bigint not null,
	application_name varchar(100) not null,
	acknowledged boolean not null,
	f_cnt bigint not null,
	tags hstore not null
);

-- NOTE: These are example indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_ack_received_at on device_ack(received_at);
create index idx_device_ack_dev_eui on device_ack(dev_eui);
create index idx_device_ack_application_id on device_ack(application_id);
create index idx_device_ack_tags on device_ack(tags);
{{</highlight>}}

### Error

Error notifications are written into the table `device_error`. The following
schema must exist:

{{<highlight sql>}}
create table device_error (
	id uuid primary key,
	received_at timestamp with time zone not null,
	dev_eui bytea not null,
	device_name varchar(100) not null,
	application_id bigint not null,
	application_name varchar(100) not null,
	type varchar(100) not null,
	error text not null,
	f_cnt bigint not null,
	tags hstore not null
);

-- NOTE: These are example indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_error_received_at on device_error(received_at);
create index idx_device_error_dev_eui on device_error(dev_eui);
create index idx_device_error_application_id on device_error(application_id);
create index idx_device_error_tags on device_error(tags);
{{</highlight>}}

### Location

Location notifications are written into the table `device_location`. The
following schema must exist:

{{<highlight sql>}}
create table device_location (
	id uuid primary key,
	received_at timestamp with time zone not null,
	dev_eui bytea not null,
	device_name varchar(100) not null,
	application_id bigint not null,
	application_name varchar(100) not null,
	altitude double precision not null,
	latitude double precision not null,
	longitude double precision not null,
	geohash varchar(12) not null,
	tags hstore not null,

	-- this field is currently not populated
	accuracy smallint not null
);

-- NOTE: These are example indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_location_received_at on device_location(received_at);
create index idx_device_location_dev_eui on device_location(dev_eui);
create index idx_device_location_application_id on device_location(application_id);
create index idx_device_location_tags on device_location(tags);
{{</highlight>}}
