alter table network_server
    drop column ca_cert,
    drop column tls_cert,
    drop column tls_key,
    drop column routing_profile_ca_cert,
    drop column routing_profile_tls_cert,
    drop column routing_profile_tls_key;