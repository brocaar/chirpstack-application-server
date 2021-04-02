alter table organization_user
    add column is_device_admin boolean not null default false,
    add column is_gateway_admin boolean not null default false;

alter table organization_user
    alter column is_device_admin drop default,
    alter column is_gateway_admin drop default;