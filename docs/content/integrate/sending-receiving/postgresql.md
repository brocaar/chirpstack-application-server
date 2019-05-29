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

* LoRa App Server will not create these tables for you. Create statements are
  given below.
* This database does not have to be the same database as used by
  LoRa App Server.
* This may generate a lot of data depending the number of devices and number
  of messages sent per device.

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
	data bytea not null,
	rx_info jsonb not null,
	object jsonb not null
);

-- NOTE: These are recommended indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_up_received_at on device_up(received_at);
create index idx_device_up_dev_eui on device_up(dev_eui);
create index idx_device_up_application_id on device_up(application_id);
create index idx_device_up_frequency on device_up(frequency);
create index idx_device_up_dr on device_up(dr);
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
	battery_level numeric(5, 2) not null
);

-- NOTE: These are recommended indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_status_received_at on device_status(received_at);
create index idx_device_status_dev_eui on device_status(dev_eui);
create index idx_device_status_application_id on device_status(application_id);
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
	dev_addr bytea not null
);

-- NOTE: These are recommended indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_join_received_at on device_join(received_at);
create index idx_device_join_dev_eui on device_join(dev_eui);
create index idx_device_join_application_id on device_join(application_id);
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
	f_cnt bigint not null
);

-- NOTE: These are recommended indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_ack_received_at on device_ack(received_at);
create index idx_device_ack_dev_eui on device_ack(dev_eui);
create index idx_device_ack_application_id on device_ack(application_id);
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
	f_cnt bigint not null
);

-- NOTE: These are recommended indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_error_received_at on device_error(received_at);
create index idx_device_error_dev_eui on device_error(dev_eui);
create index idx_device_error_application_id on device_error(application_id);
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

	-- this field is currently not populated
	accuracy smallint not null
);

-- NOTE: These are recommended indices, depending on how this table is being
-- used, you might want to change these.
create index idx_device_location_received_at on device_location(received_at);
create index idx_device_location_dev_eui on device_location(dev_eui);
create index idx_device_location_application_id on device_location(application_id);
{{</highlight>}}
