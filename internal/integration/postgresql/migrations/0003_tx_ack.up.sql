CREATE TABLE device_txack (
    id               uuid primary key,
    received_at      timestamp with time zone not null,
    dev_eui          bytea not null,
    device_name      varchar(100) not null,
    application_id   bigint not null,
    application_name varchar(100) not null,
    gateway_id       bytea not null,
    f_cnt            bigint not null,
    tags             hstore not null,
    tx_info          jsonb not null
);
