-- +migrate Up
alter table node
	add column app_s_key bytea not null,
	add column nwk_s_key bytea not null;

-- +migrate Down
alter table node
	drop column app_s_key,
	drop column nwk_s_key;
