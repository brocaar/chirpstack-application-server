-- +migrate Up
create table "user" (
	id bigserial primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	username character varying (100) not null,
	password_hash character varying (200) not null,
	session_ttl bigint not null,
	is_active boolean not null,
	is_admin boolean not null
);

create unique index idx_user_username on "user"(username);

create table application_user (
	id bigserial primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	user_id bigint not null,
	application_id bigint not null,
	is_admin boolean not null,

	unique(user_id, application_id)
);

create index idx_application_user_user_id on application_user(user_id);
create index idx_application_user_application_id on application_user(application_id);


-- +migrate Down
drop index idx_application_user_application_id;
drop index idx_application_user_user_id;
drop table application_user;

drop index idx_user_username;
drop table "user";
