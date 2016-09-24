-- +migrate Up
alter table node
	add column app_s_key bytea not null,
	add column nwk_s_key bytea not null,
	add column dev_addr bytea not null,
	add column name varchar(100) not null default '';

-- +migrate Down
alter table node
	drop column app_s_key,
	drop column nwk_s_key,
	drop column dev_addr,
	drop column name;
