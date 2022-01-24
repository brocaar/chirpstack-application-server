drop index idx_node_name;

update node set name = description;

alter table node
	drop column description,
	drop column is_abp,
	drop constraint node_application_id_name_key;