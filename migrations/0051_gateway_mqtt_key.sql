-- +migrate Up
alter table gateway
    add column mqtt_key bytea,
    add column mqtt_key_hash character varying (200);

-- +migrate Down
alter table gateway
    drop column mqtt_key,
    drop column mqtt_key_hash;
