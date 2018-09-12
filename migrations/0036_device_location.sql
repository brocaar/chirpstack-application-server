-- +migrate Up
alter table device
    add column latitude double precision,
    add column longitude double precision,
    add column altitude double precision;

-- +migrate Down
alter table device
    drop column latitude,
    drop column longitude,
    drop column altitude;
