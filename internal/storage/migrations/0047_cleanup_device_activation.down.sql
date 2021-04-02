create table device_activation (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    dev_eui bytea not null references device on delete cascade,
    dev_addr bytea not null,
    app_s_key bytea not null
);

create index idx_device_activation_dev_eui on device_activation(dev_eui);