alter table gateway
    add column tags hstore,
    add column metadata hstore;

create index idx_gateway_tags on gateway using gin (tags);