drop index idx_node_application_id;

alter table node
	drop column application_id;

drop index idx_application_name;
drop table application;