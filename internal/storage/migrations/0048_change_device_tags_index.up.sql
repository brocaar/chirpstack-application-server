drop index idx_device_tags;
create index idx_device_tags on device using gin (tags);