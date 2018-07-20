-- +migrate Up
alter table device_activation
    drop column f_nwk_s_int_key,
    drop column s_nwk_s_int_key,
    drop column nwk_s_enc_key;

-- +migrate Down
alter table device_activation
    add column f_nwk_s_int_key bytea,
    add column s_nwk_s_int_key bytea,
    add column nwk_s_enc_key bytea;
