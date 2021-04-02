create index idx_service_profile_updated_at on service_profile(updated_at);
create index idx_service_profile_created_at on service_profile(created_at);

create index idx_network_server_updated_at on network_server(updated_at);
create index idx_network_server_created_at on network_server(created_at);

create index idx_gateway_profile_updated_at on gateway_profile(updated_at);
create index idx_gateway_profile_created_at on gateway_profile(created_at);

create index idx_gateway_ping_rx_created_at on gateway_ping_rx(created_at);

create index idx_gateway_ping_created_at on gateway_ping(created_at);

create index device_queue_mapping_created_at on device_queue_mapping(created_at);

create index idx_device_profile_updated_at on device_profile(updated_at);
create index idx_device_profile_created_at on device_profile(created_at);

create index idx_device_keys_updated_at on device_keys(updated_at);
create index idx_device_keys_created_at on device_keys(created_at);

create index idx_device_activation_created_at on device_activation(created_at);

create index idx_device_updated_at on device(updated_at);
create index idx_device_created_at on device(created_at);