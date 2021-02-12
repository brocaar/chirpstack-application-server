CREATE EXTENSION IF NOT EXISTS hstore;

CREATE TABLE IF NOT EXISTS device_up (
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

CREATE INDEX IF NOT EXISTS idx_device_up_received_at on device_up(received_at);
CREATE INDEX IF NOT EXISTS idx_device_up_dev_eui on device_up(dev_eui);
CREATE INDEX IF NOT EXISTS idx_device_up_application_id on device_up(application_id);
CREATE INDEX IF NOT EXISTS idx_device_up_frequency on device_up(frequency);
CREATE INDEX IF NOT EXISTS idx_device_up_dr on device_up(dr);
CREATE INDEX IF NOT EXISTS idx_device_up_tags on device_up(tags);

CREATE TABLE IF NOT EXISTS device_status (
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

CREATE INDEX IF NOT EXISTS idx_device_status_received_at on device_status(received_at);
CREATE INDEX IF NOT EXISTS idx_device_status_dev_eui on device_status(dev_eui);
CREATE INDEX IF NOT EXISTS idx_device_status_application_id on device_status(application_id);
CREATE INDEX IF NOT EXISTS idx_device_status_tags on device_status(tags);

CREATE TABLE IF NOT EXISTS device_join (
    id uuid primary key,
    received_at timestamp with time zone not null,
    dev_eui bytea not null,
    device_name varchar(100) not null,
    application_id bigint not null,
    application_name varchar(100) not null,
    dev_addr bytea not null,
    tags hstore not null
);

CREATE INDEX IF NOT EXISTS idx_device_join_received_at on device_join(received_at);
CREATE INDEX IF NOT EXISTS idx_device_join_dev_eui on device_join(dev_eui);
CREATE INDEX IF NOT EXISTS idx_device_join_application_id on device_join(application_id);
CREATE INDEX IF NOT EXISTS idx_device_join_tags on device_join(tags);

CREATE TABLE IF NOT EXISTS device_ack (
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

CREATE INDEX IF NOT EXISTS idx_device_ack_received_at on device_ack(received_at);
CREATE INDEX IF NOT EXISTS idx_device_ack_dev_eui on device_ack(dev_eui);
CREATE INDEX IF NOT EXISTS idx_device_ack_application_id on device_ack(application_id);
CREATE INDEX IF NOT EXISTS idx_device_ack_tags on device_ack(tags);

CREATE TABLE IF NOT EXISTS device_error (
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

CREATE INDEX IF NOT EXISTS idx_device_error_received_at on device_error(received_at);
CREATE INDEX IF NOT EXISTS idx_device_error_dev_eui on device_error(dev_eui);
CREATE INDEX IF NOT EXISTS idx_device_error_application_id on device_error(application_id);
CREATE INDEX IF NOT EXISTS idx_device_error_tags on device_error(tags);

CREATE TABLE IF NOT EXISTS device_location (
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

CREATE INDEX IF NOT EXISTS idx_device_location_received_at on device_location(received_at);
CREATE INDEX IF NOT EXISTS idx_device_location_dev_eui on device_location(dev_eui);
CREATE INDEX IF NOT EXISTS idx_device_location_application_id on device_location(application_id);
CREATE INDEX IF NOT EXISTS idx_device_location_tags on device_location(tags);
