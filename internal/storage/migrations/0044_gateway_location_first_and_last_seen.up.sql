alter table gateway
    add column first_seen_at timestamp with time zone,
    add column last_seen_at timestamp with time zone,
    add column latitude double precision not null default 0,
    add column longitude double precision not null default 0,
    add column altitude double precision not null default 0;

alter table gateway
    alter column latitude drop default,
    alter column longitude drop default,
    alter column altitude drop default;