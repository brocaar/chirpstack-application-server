ALTER TABLE IF EXISTS device_up
    ADD COLUMN IF NOT EXISTS dev_addr bytea,
    ADD COLUMN IF NOT EXISTS confirmed_uplink boolean not null default false;