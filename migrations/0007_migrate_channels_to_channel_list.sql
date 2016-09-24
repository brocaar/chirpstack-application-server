-- +migrate Up
alter table channel_list
	add column channels integer[];

update channel_list
	set channels=(
		select array(
			select frequency
			from
				channel
			where
				channel_list_id=channel_list.id
			order by
				channel));

drop table channel;

-- +migrate Down
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
