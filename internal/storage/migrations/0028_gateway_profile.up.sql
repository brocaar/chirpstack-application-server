create table gateway_profile (
    gateway_profile_id uuid primary key,
    network_server_id bigint not null references network_server,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    name varchar(100) not null
);

create index idx_gateway_profile_network_server_id on gateway_profile(network_server_id);
create index idx_gateway_profile_created_at on gateway_profile(created_at);
create index idx_gateway_profile_updated_at on gateway_profile(updated_at);

alter table gateway
    add column gateway_profile_id uuid references gateway_profile;

create index idx_gateway_gateway_profile_id on gateway(gateway_profile_id);