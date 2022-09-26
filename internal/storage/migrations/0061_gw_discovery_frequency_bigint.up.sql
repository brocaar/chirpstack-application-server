alter table network_server
    alter column gateway_discovery_tx_frequency type bigint;
alter table gateway_ping
    alter column frequency type bigint;
