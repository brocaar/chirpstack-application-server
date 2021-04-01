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