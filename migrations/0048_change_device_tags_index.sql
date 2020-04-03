-- +migrate Up
drop index idx_device_tags;
create index idx_device_tags on device using gin (tags);

-- +migrate Down
drop index idx_device_tags;
create index idx_device_tags on device using btree (tags);
