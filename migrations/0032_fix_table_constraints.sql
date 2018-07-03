-- +migrate Up
alter table application
    alter column service_profile_id set not null;

create unique index idx_device_name_application_id on device(name, application_id);

alter table gateway
    alter column network_server_id set not null;

create unique index idx_gateway_name_organization_id on gateway(name, organization_id);

-- +migrate Down
drop index idx_gateway_name_organization_id;

alter table gateway
    alter column network_server_id drop not null;

drop index idx_device_name_application_id;

alter table application
    alter column service_profile_id drop not null;
