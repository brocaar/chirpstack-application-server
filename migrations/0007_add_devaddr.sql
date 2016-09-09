-- +migrate Up
alter table node
	add column dev_addr bytea not null;

-- +migrate Down
alter table node
	drop column dev_addr;
