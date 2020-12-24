-- +migrate Up
alter table gateway
    add column service_profile_id uuid references service_profile;

    create index idx_gateway_service_profile_id on gateway(service_profile_id);

-- +migrate Down
drop index idx_gateway_service_profile_id;

alter table gateway
    drop column service_profile_id;
