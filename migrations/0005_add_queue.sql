-- +migrate Up
create table downlink_queue (
    id bigserial,
	reference varchar(100) not null,
    dev_eui bytea references node on delete cascade not null,
    confirmed boolean not null default false,
    pending boolean not null default false,
    fport smallint not null,
    data bytea not null,
    date_created timestamp with time zone not null default now()
);

create index downlink_queue_dev_eui on downlink_queue(dev_eui);

-- +migrate Down
drop index downlink_queue_dev_eui;

drop table downlink_queue;
