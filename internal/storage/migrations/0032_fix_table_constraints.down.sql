drop index idx_gateway_name_organization_id;

alter table gateway
    alter column network_server_id drop not null;

drop index idx_device_name_application_id;

alter table application
    alter column service_profile_id drop not null;