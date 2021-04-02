alter table network_server
    add column ca_cert text not null default '',
    add column tls_cert text not null default '',
    add column tls_key text not null default '',
    add column routing_profile_ca_cert text not null default '',
    add column routing_profile_tls_cert text not null default '',
    add column routing_profile_tls_key text not null default '';