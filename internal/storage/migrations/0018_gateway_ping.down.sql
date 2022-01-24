drop index idx_gateway_last_ping_sent_at;
drop index idx_gateway_ping;
alter table gateway
    drop column ping,
    drop column last_ping_id,
    drop column last_ping_sent_at;

drop index idx_gateway_ping_rx_gateway_mac;
drop index idx_gateway_ping_rx_ping_id;
drop index idx_gateway_ping_rx_created_at;
drop table gateway_ping_rx;

drop index idx_gateway_ping_gateway_mac;
drop index idx_gateway_ping_created_at;
drop table gateway_ping;