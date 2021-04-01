drop index idx_gateway_service_profile_id;

alter table gateway
    drop column service_profile_id;