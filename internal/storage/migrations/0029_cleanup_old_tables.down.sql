create table node (
    dev_eui bytea NOT NULL,
    app_eui bytea NOT NULL,
    app_key bytea NOT NULL,
    used_dev_nonces bytea,
    rx_delay smallint NOT NULL,
    rx1_dr_offset smallint NOT NULL,
    rx_window smallint DEFAULT 0 NOT NULL,
    rx2_dr smallint DEFAULT 0 NOT NULL,
    app_s_key bytea DEFAULT '\x00000000000000000000000000000000'::bytea NOT NULL,
    nwk_s_key bytea DEFAULT '\x00000000000000000000000000000000'::bytea NOT NULL,
    dev_addr bytea DEFAULT '\x00000000'::bytea NOT NULL,
    name character varying(100) DEFAULT ''::character varying NOT NULL,
    relax_fcnt boolean DEFAULT false NOT NULL,
    adr_interval integer DEFAULT 0 NOT NULL,
    installation_margin numeric(5,2) DEFAULT 0 NOT NULL,
    application_id bigint NOT NULL,
    description text NOT NULL,
    is_abp boolean DEFAULT false NOT NULL,
    is_class_c boolean DEFAULT false NOT NULL,
    use_application_settings boolean DEFAULT false NOT NULL
);

alter table only node
    add constraint node_application_id_name_key unique (application_id, name);

alter table only node
    add constraint node_pkey primary key (dev_eui);

create index idx_node_application_id on node using btree (application_id);
create index idx_node_dev_eui_prefix on node using btree (encode(dev_eui, 'hex'::text) varchar_pattern_ops);
create index idx_node_name on node using btree (name);
create index idx_node_name_prefix on node using btree (name varchar_pattern_ops);
create index node_app_eui on node using btree (app_eui);


create table downlink_queue (
    id bigserial,
	reference varchar(100) not null,
    dev_eui bytea references node on delete cascade not null,
    confirmed boolean not null default false,
    pending boolean not null default false,
    fport smallint not null,
    data bytea not null
);

create index downlink_queue_dev_eui on downlink_queue(dev_eui);


create table application_user (
	id bigserial primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	user_id bigint not null references "user" on delete cascade,
	application_id bigint not null references application on delete cascade,
	is_admin boolean not null,

	unique(user_id, application_id)
);

create index idx_application_user_user_id on application_user(user_id);
create index idx_application_user_application_id on application_user(application_id);


create table device_queue (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    reference varchar(100) not null,
    dev_eui bytea not null references device on delete cascade,
    confirmed boolean not null default false,
    pending boolean not null default false,
    fport smallint not null,
    data bytea not null
);

create index idx_device_queue_created_at on device_queue(created_at);
create index idx_device_queue_updated_at on device_queue(updated_at);
create index idx_device_queue_dev_eui on device_queue(dev_eui);