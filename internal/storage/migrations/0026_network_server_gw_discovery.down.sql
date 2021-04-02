alter table network_server
    drop column gateway_discovery_enabled,
    drop column gateway_discovery_interval,
    drop column gateway_discovery_tx_frequency,
    drop column gateway_discovery_dr;