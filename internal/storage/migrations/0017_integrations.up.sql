create table integration (
	id bigserial primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	application_id bigint not null references application on delete cascade,
	kind character varying (20) not null,
	settings jsonb,

	constraint integration_kind_application_id unique (kind, application_id)
);

create index idx_integration_kind on integration(kind);
create index idx_integration_application_id on integration(application_id);