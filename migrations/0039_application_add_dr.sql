-- +migrate Up
alter table device
    add column dr smallint;

-- +migrate Down
alter table device
    drop column dr;

