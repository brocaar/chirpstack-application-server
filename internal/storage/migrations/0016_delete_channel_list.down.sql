create table channel_list (
	id bigserial primary key,
	name character varying (100) not null,
	channels integer[]
);

alter table node
	add column channel_list_id bigint references channel_list on delete set null;

alter table application
	add column channel_list_id bigint references channel_list on delete set null;