create table device_queue_mapping (
    id bigserial primary key,
    created_at timestamp with time zone not null,
    reference text not null,
    dev_eui bytea references device on delete cascade not null,
    f_cnt int not null
);

create index device_queue_mapping_created_at on device_queue_mapping(created_at);
create index device_queue_mapping_dev_eui on device_queue_mapping(dev_eui);