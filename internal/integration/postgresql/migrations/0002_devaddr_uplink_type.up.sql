ALTER TABLE device_up
    ADD COLUMN dev_addr bytea,
    ADD COLUMN confirmed_uplink boolean not null default false;
