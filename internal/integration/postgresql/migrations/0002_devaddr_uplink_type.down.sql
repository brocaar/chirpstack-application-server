ALTER TABLE IF EXISTS device_up
    DROP COLUMN IF EXISTS dev_addr,
    DROP COLUMN IF EXISTS confirmed_uplink;