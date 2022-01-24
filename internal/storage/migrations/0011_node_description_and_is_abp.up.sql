alter table node
	add column description text,
	add column is_abp boolean not null default false;

update node set description = name;
update node set name = encode(dev_eui, 'hex');

alter table node
	alter column description set not null,
	add constraint node_application_id_name_key unique (application_id, name);

create index idx_node_name on node(name);
