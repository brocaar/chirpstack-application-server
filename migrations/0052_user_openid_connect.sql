-- +migrate Up
drop index idx_user_username;
drop index idx_user_username_trgm;

alter table "user"
    rename column email to email_old;

alter table "user"
    rename column username to email;

alter table "user"
    add column external_id text null,
    add column email_verified bool not null default false;

alter table "user"
    alter column email_verified drop default,
    alter column "note" drop default,
    alter column email type text;

create unique index idx_user_email on "user" (email);
create unique index idx_user_external_id on "user" (external_id);

-- +migrate Down
drop index idx_user_external_id;
drop index idx_user_email;

alter table "user"
    alter column email type varchar(100),
    alter column "note" set default '';

alter table "user"
    drop column email_verified,
    drop column external_id;

alter table "user"
    rename column email to username;

alter table "user"
    rename column email_old to email;

create index idx_user_username_trgm on "user" using gin (username gin_trgm_ops);
create unique index idx_user_username on "user" (username);
