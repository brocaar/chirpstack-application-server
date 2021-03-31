alter table device_up
    add column tx_info jsonb not null default 'null';

alter table device_up
    alter column tx_info drop default;
