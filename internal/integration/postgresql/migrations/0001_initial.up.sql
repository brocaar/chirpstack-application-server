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