-- +migrate Up
create table downlink_queue (
    id bigserial,
    dev_eui bytea references node on delete cascade not null,
    confirmed boolean,
    pending boolean,
    fport smallint,
    data bytea
);

create index downlink_queue_dev_eui on downlink_queue(dev_eui);

-- +migrate Down
drop index downlink_queue_dev_eui;

drop table downlink_queue;
