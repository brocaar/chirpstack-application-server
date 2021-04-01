drop index idx_application_name;
drop index idx_organization_display_name_prefix;
drop index idx_device_dev_eui_prefix;
drop index idx_device_name_prefix;
drop index idx_user_username_prefix;

create index idx_application_name_trgm on application using gin (name gin_trgm_ops);

create index idx_gateway_mac_trgm on gateway using gin (encode(mac, 'hex') gin_trgm_ops);
create index idx_gateway_name_trgm on gateway using gin (name gin_trgm_ops);

create index idx_device_dev_eui_trgm on device using gin (encode(dev_eui, 'hex') gin_trgm_ops);
create index idx_device_name_trgm on device using gin (name gin_trgm_ops);

create index idx_organization_name_trgm on organization using gin (name gin_trgm_ops);

create index idx_user_username_trgm on "user" using gin (username gin_trgm_ops);