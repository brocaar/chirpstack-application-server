create table gateway_ping (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    gateway_mac bytea not null references gateway on delete cascade,
    frequency integer not null,
    dr integer not null
);

create index idx_gateway_ping_created_at on gateway_ping(created_at);
create index idx_gateway_ping_gateway_mac on gateway_ping(gateway_mac);

create table gateway_ping_rx (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    ping_id bigint not null references gateway_ping on delete cascade,
    gateway_mac bytea not null references gateway on delete cascade,
    received_at timestamp with time zone,
    rssi integer not null,
    lora_snr decimal(3,1) not null,
    location point,
    altitude double precision
);

create index idx_gateway_ping_rx_created_at on gateway_ping_rx(created_at);
create index idx_gateway_ping_rx_ping_id on gateway_ping_rx(ping_id);
create index idx_gateway_ping_rx_gateway_mac on gateway_ping_rx(gateway_mac);

alter table gateway
    add column ping boolean not null default false,
    add column last_ping_id bigint references gateway_ping on delete set null,
    add column last_ping_sent_at timestamp with time zone;

create index idx_gateway_ping on gateway(ping);
create index idx_gateway_last_ping_sent_at on gateway(last_ping_sent_at);