-- +migrate Up
alter table "user"
    add column email text not null default '',
    add column note text not null default '';

-- +migrate Down
alter table "user"
    drop column email,
    drop column note;
