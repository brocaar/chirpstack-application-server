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
