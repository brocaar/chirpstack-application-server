alter table channel_list
	drop column channels;

create table channel (
	id bigserial primary key,
	channel_list_id bigint references channel_list on delete cascade not null,
	channel integer not null,
	frequency integer not null,
	check (channel >= 3 and channel <= 7 and frequency > 0),
	unique (channel_list_id, channel)
);