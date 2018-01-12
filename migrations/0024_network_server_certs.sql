-- +migrate Up
alter table network_server
    add column ca_cert text not null default '',
    add column tls_cert text not null default '',
    add column tls_key text not null default '',
    add column routing_profile_ca_cert text not null default '',
    add column routing_profile_tls_cert text not null default '',
    add column routing_profile_tls_key text not null default '';

-- +migrate Down
alter table network_server
    drop column ca_cert,
    drop column tls_cert,
    drop column tls_key,
    drop column routing_profile_ca_cert,
    drop column routing_profile_tls_cert,
    drop column routing_profile_tls_key;
