-- +migrate Up
create table api_key (
    id uuid primary key,
    created_at timestamp with time zone not null,
    name varchar(100) not null,
    is_admin boolean not null default false,
    organization_id bigint references organization on delete cascade,
    application_id bigint references application on delete cascade
);

create index idx_api_key_organization_id on api_key(organization_id);
create index idx_api_key_application_id on api_key(application_id);

-- +migrate Down
drop index idx_api_key_application_id;
drop index idx_api_key_organization_id;
drop table api_key;
