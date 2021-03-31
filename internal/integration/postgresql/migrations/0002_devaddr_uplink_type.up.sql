ALTER TABLE device_up
    ADD COLUMN dev_addr bytea not null default '',
    ADD COLUMN confirmed_uplink boolean not null default false;

ALTER TABLE device_up
    ALTER COLUMN dev_addr drop default,
    ALTER COLUMN confirmed_uplink drop default;
