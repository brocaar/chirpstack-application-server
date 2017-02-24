-- +migrate Up
alter table node
	add column is_class_c boolean not null default false;

-- +migrate Down
alter table node
	drop column is_class_c;
