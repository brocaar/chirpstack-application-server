drop index idx_user_username_trgm;
drop index idx_organization_name_trgm;
drop index idx_device_name_trgm;
drop index idx_device_dev_eui_trgm;
drop index idx_gateway_name_trgm;
drop index idx_gateway_mac_trgm;
drop index idx_application_name_trgm;

create index idx_user_username_prefix on "user"(username varchar_pattern_ops);

create index idx_device_dev_eui_prefix on device(encode(dev_eui, 'hex') varchar_pattern_ops);
create index idx_device_name_prefix on device(name varchar_pattern_ops);

create index idx_organization_display_name_prefix on organization(lower(display_name) varchar_pattern_ops);

create index idx_application_name on application(name);