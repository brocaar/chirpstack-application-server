create table network_server (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    name varchar(100) not null,
    server varchar(255) not null 
);

create index idx_network_server_created_at on network_server(created_at);
create index idx_network_server_updated_at on network_server(updated_at);

create table service_profile (
    service_profile_id uuid primary key,
    organization_id bigint not null references organization,
    network_server_id bigint not null references network_server,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    name varchar(100) not null
);

create index idx_service_profile_organization_id on service_profile(organization_id);
create index idx_service_profile_network_server_id on service_profile(network_server_id);
create index idx_service_profile_created_at on service_profile(created_at);
create index idx_service_profile_updated_at on service_profile(updated_at);

create table device_profile (
    device_profile_id uuid primary key,
    network_server_id bigint not null references network_server,
    organization_id bigint not null references organization,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    name varchar(100) not null
);

create index idx_device_profile_network_server_id on device_profile(network_server_id);
create index idx_device_profile_organization_id on device_profile(organization_id);
create index idx_device_profile_created_at on device_profile(created_at);
create index idx_device_profile_updated_at on device_profile(updated_at);

create table device (
    dev_eui bytea primary key,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    application_id bigint not null references application,
    device_profile_id uuid not null references device_profile,
    name varchar(100) not null,
    description text not null
);

create index idx_device_created_at on device(created_at);
create index idx_device_updated_at on device(updated_at);
create index idx_device_application_id on device(application_id);
create index idx_device_device_profile_id on device(device_profile_id);
create index idx_device_dev_eui_prefix on device(encode(dev_eui, 'hex') varchar_pattern_ops);
create index idx_device_name_prefix on device(name varchar_pattern_ops);

create table device_keys (
    dev_eui bytea primary key references device on delete cascade,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    app_key bytea not null,
    join_nonce integer not null
);

create index idx_device_keys_created_at on device_keys(created_at);
create index idx_device_keys_updated_at on device_keys(updated_at);

create table device_activation (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    dev_eui bytea not null references device on delete cascade,
    dev_addr bytea not null,
    app_s_key bytea not null,
    nwk_s_key bytea not null
);

create index idx_device_activation_created_at on device_activation(created_at);
create index idx_device_activation_dev_eui on device_activation(dev_eui);

create table device_queue (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    reference varchar(100) not null,
    dev_eui bytea not null references device on delete cascade,
    confirmed boolean not null default false,
    pending boolean not null default false,
    fport smallint not null,
    data bytea not null
);

create index idx_device_queue_created_at on device_queue(created_at);
create index idx_device_queue_updated_at on device_queue(updated_at);
create index idx_device_queue_dev_eui on device_queue(dev_eui);

alter table application
    add column service_profile_id uuid references service_profile,
    drop column rx_delay,
    drop column rx1_dr_offset,
    drop column rx_window,
    drop column rx2_dr,
    drop column relax_fcnt,
    drop column adr_interval,
    drop column installation_margin,
    drop column is_abp,
    drop column is_class_c;

create index idx_application_service_profile_id on application(service_profile_id);

alter table gateway
    add column network_server_id bigint references network_server;

create index idx_gateway_network_server_id on gateway(network_server_id);