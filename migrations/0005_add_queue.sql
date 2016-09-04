-- +migrate Up
create table node_queue (
    id bigserial,
    dev_eui bytea references node on delete cascade not null,
    confirmed boolean,
    pending boolean,
    fport smallint,
    data bytea
);

create index node_queue_dev_eui on node_queue(dev_eui);

-- +migrate Down
drop index node_queue_dev_eui;

drop table node_queue;
