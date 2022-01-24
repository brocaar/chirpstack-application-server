alter table node
	add column app_s_key bytea not null default E'\\x00000000000000000000000000000000',
	add column nwk_s_key bytea not null default E'\\x00000000000000000000000000000000',
	add column dev_addr bytea not null default E'\\x00000000',
	add column name varchar(100) not null default '';

update node set name=(select name from application where app_eui = node.app_eui);
