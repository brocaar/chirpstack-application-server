create table application (
	id bigserial primary key,
	name varchar(100) not null,
	description text not null,

	constraint application_name_key unique (name)
);

create index idx_application_name on application(name);

insert into application
	(name, description)
	select distinct(encode(app_eui, 'hex')) as name, 'Application ' || encode(app_eui, 'hex') as description from node;

alter table node 
	add column application_id bigint references application on delete cascade;

update node set application_id = (select id from application where name = encode(node.app_eui, 'hex'));

alter table node 
	alter column application_id set not null;

create index idx_node_application_id on node(application_id);
