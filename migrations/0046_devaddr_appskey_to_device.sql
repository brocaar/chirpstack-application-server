-- +migrate Up
alter table device
    add column dev_addr bytea,
    add column app_s_key bytea;

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

-- +migrate Down
alter table device
    drop column app_s_key,
    drop column dev_addr;
