-- +migrate Up
alter table organization
    add column max_device_count integer not null default 0,
    add column max_gateway_count integer not null default 0;

alter table organization
    alter column max_device_count drop default,
    alter column max_gateway_count drop default;

-- +migrate Down
alter table organization
    drop column max_device_count,
    drop column max_gateway_count;
