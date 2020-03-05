-- +migrate Up
alter table gateway
    add column tags hstore,
    add column metadata hstore;

create index idx_gateway_tags on gateway using gin (tags);

-- +migrate Down
drop index idx_gateway_tags;

alter table gateway
    drop column metadata,
    drop column tags;
