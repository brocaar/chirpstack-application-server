-- +migrate Up
alter table device
    add column last_seen_at timestamp with time zone null,
    add column device_status_battery int null,
    add column device_status_margin int null;

-- +migrate Down
alter table device
    drop column last_seen_at,
    drop column device_status_battery,
    drop column device_status_margin;
