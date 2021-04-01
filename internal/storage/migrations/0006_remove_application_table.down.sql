create table application (
    app_eui bytea primary key,
    name character varying (100) not null
);

insert into application
    select distinct(app_eui), 'no name' as name from node;

alter table node
    add constraint node_app_eui_fkey foreign key(app_eui) references application on delete cascade;