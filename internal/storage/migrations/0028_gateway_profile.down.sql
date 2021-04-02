drop index idx_gateway_gateway_profile_id;
alter table gateway
    drop column gateway_profile_id;

drop index idx_gateway_profile_network_server_id;
drop index idx_gateway_profile_created_at;
drop index idx_gateway_profile_updated_at;

drop table gateway_profile;