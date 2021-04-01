drop index idx_application_organization_id;
alter table application
	drop constraint application_name_organization_id_key,
	add constraint application_name_key unique (name),
	drop column organization_id;

drop index idx_gateway_organization_id;
drop table gateway;

drop index idx_organization_user_organization_id;
drop index idx_organization_user_user_id;
drop table organization_user;

drop index idx_organization_display_name_prefix;
drop index idx_organization_name;
drop table organization;