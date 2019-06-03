-- +migrate Up
alter table device
    add column variables hstore,
    add column tags hstore;

create index idx_device_tags on device(tags);

-- +migrate Down
drop index idx_device_tags;

alter table device
    drop column variables,
    drop column tags;

