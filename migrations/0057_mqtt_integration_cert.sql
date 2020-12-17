-- +migrate Up
alter table application
    add column mqtt_tls_cert bytea;

-- +migrate Down
alter table application
    drop column mqtt_tls_cert;
