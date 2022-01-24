alter table device_profile
    add column tags hstore;

create index idx_device_profile_tags on device_profile using gin (tags);