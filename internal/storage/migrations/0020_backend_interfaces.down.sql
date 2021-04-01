drop index idx_gateway_network_server_id;
alter table gateway
    drop column network_server_id;

drop index idx_application_service_profile_id;
alter table application
    drop column service_profile_id,
    add column rx_delay int2 not null default 0,
    add column rx1_dr_offset int2 not null default 0,
    add column rx_window int2 not null default 0,
    add column rx2_dr int2 not null default 0,
    add column relax_fcnt boolean not null default false,
    add column adr_interval integer not null default 0,
    add column installation_margin decimal(5,2) not null default 0,
    add column is_abp boolean not null default false,
    add column is_class_c boolean not null default false;

drop index idx_device_queue_dev_eui;
drop index idx_device_queue_updated_at;
drop index idx_device_queue_created_at;
drop table device_queue;

drop index idx_device_activation_dev_eui;
drop index idx_device_activation_created_at;
drop table device_activation;

drop index idx_device_keys_updated_at;
drop index idx_device_keys_created_at;
drop table device_keys;

drop index idx_device_dev_eui_prefix;
drop index idx_device_name_prefix;
drop index idx_device_application_id;
drop index idx_device_device_profile_id;
drop index idx_device_updated_at;
drop index idx_device_created_at;
drop table device;

drop index idx_device_profile_updated_at;
drop index idx_device_profile_created_at;
drop index idx_device_profile_organization_id;
drop index idx_device_profile_network_server_id;
drop table device_profile;

drop index idx_service_profile_updated_at;
drop index idx_service_profile_created_at;
drop index idx_service_profile_network_server_id;
drop index idx_service_profile_organization_id;
drop table service_profile;

drop index idx_network_server_updated_at;
drop index idx_network_server_created_at;
drop table network_server;