alter table device
    add column variables hstore,
    add column tags hstore;

create index idx_device_tags on device(tags);