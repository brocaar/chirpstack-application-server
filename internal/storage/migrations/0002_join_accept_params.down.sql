alter table node
	drop column rx_delay,
	drop column rx1_dr_offset,
	drop column channel_list_id;

drop table channel;

drop table channel_list;
