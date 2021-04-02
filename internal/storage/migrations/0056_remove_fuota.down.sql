alter table device_keys
    add column gen_app_key bytea not null default decode('00000000000000000000000000000000', 'hex');

alter table device_keys
    alter column gen_app_key drop default;

alter table multicast_group
    add column mc_key bytea not null default decode('00000000000000000000000000000000', 'hex');

alter table multicast_group
    alter column mc_key drop default;

create table remote_multicast_setup (
    dev_eui bytea not null references device on delete cascade,
    multicast_group_id uuid not null references multicast_group on delete cascade,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    mc_group_id smallint not null,
    mc_addr bytea not null,
    mc_key_encrypted bytea not null,
    min_mc_f_cnt bigint not null,
    max_mc_f_cnt bigint not null,
    state varchar(20) not null,
    state_provisioned bool not null default false,
    retry_after timestamp with time zone not null,
    retry_count smallint not null,
    retry_interval bigint not null,

    primary key(dev_eui, multicast_group_id)
);

create index idx_remote_multicast_setup_state_provisioned on remote_multicast_setup(state_provisioned);
create index idx_remote_multicast_setup_retry_after on remote_multicast_setup(retry_after);

create table remote_multicast_class_c_session (
    dev_eui bytea not null references device on delete cascade,
    multicast_group_id uuid not null references multicast_group on delete cascade,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    mc_group_id smallint not null,
    session_time timestamp with time zone not null,
    session_time_out smallint not null,
    dl_frequency integer not null,
    dr smallint not null,
    state_provisioned bool not null default false,
    retry_after timestamp with time zone not null,
    retry_count smallint not null,
    retry_interval bigint not null,

    primary key(dev_eui, multicast_group_id)
);

create index idx_remote_multicast_class_c_session_state_provisioned on remote_multicast_class_c_session(state_provisioned);
create index idx_remote_multicast_class_c_session_state_retry_after on remote_multicast_class_c_session(retry_after);

create table remote_fragmentation_session (
    dev_eui bytea not null references device on delete cascade,
    frag_index smallint not null,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    mc_group_ids smallint[],
    nb_frag integer not null,
    frag_size smallint not null,
    fragmentation_matrix bytea not null,
    block_ack_delay smallint not null,
    padding smallint not null,
    descriptor bytea not null,
    state varchar(20) not null,
    state_provisioned bool not null default false,
    retry_after timestamp with time zone not null,
    retry_count smallint not null,
    retry_interval bigint not null,

    primary key(dev_eui, frag_index)
);

create index idx_remote_fragmentation_session_state_provisioned on remote_fragmentation_session(state_provisioned);
create index idx_remote_fragmentation_session_retry_after on remote_fragmentation_session(retry_after);

create table fuota_deployment (
    id uuid primary key,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    name varchar(100) not null,
    multicast_group_id uuid references multicast_group on delete set null,
    group_type char not null,
    dr smallint not null,
    frequency int not null,
    ping_slot_period smallint not null,
    fragmentation_matrix bytea not null,
    descriptor bytea not null,
    payload bytea not null,
    frag_size smallint not null,
    redundancy smallint not null,
    multicast_timeout smallint not null,
    block_ack_delay smallint not null,
    state varchar(20) not null,
    unicast_timeout bigint not null,
    next_step_after timestamp with time zone not null
);

create index idx_fuota_deployment_multicast_group_id on fuota_deployment(multicast_group_id);
create index idx_fuota_deployment_state on fuota_deployment(state);
create index idx_fuota_deployment_next_step_after on fuota_deployment(next_step_after);

create table fuota_deployment_device (
    fuota_deployment_id uuid not null references fuota_deployment on delete cascade,
    dev_eui bytea not null references device on delete cascade,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    state varchar(20) not null,
    error_message text not null,

    primary key(fuota_deployment_id, dev_eui)
);