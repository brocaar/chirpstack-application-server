alter table device
    add column dev_addr bytea not null default '\x00000000',
    add column app_s_key bytea not null default '\x00000000000000000000000000000000';

update device d
    set
        dev_addr = da.dev_addr,
        app_s_key = da.app_s_key
    from
        (
            select
                distinct on (dev_eui) *
            from
                device_activation
            order by
                dev_eui,
                created_at desc
        ) da
    where
        d.dev_eui = da.dev_eui;

alter table device
    alter column dev_addr drop default,
    alter column app_s_key drop default;