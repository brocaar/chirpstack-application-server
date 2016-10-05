-- +migrate Up
alter table node
	add column relax_fcnt boolean not null default false;

-- +migrate Down
alter table node
	drop column relax_fcnt;
