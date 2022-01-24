alter table node
	drop column channel_list_id;

alter table application
	drop column channel_list_id;

drop table channel_list;