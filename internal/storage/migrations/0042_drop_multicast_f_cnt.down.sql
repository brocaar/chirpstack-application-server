alter table multicast_group
    add column f_cnt bigint not null default 0;

alter table multicast_group
    alter column f_cnt drop default;