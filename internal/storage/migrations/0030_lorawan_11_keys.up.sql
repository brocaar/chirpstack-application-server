alter table device_keys
    rename column app_key to nwk_key;

alter table device_keys
    add column app_key bytea not null default decode('00000000000000000000000000000000', 'hex');

alter table device_keys
    alter column app_key drop default;

alter table device_activation
    rename column nwk_s_key to f_nwk_s_int_key;

alter table device_activation
    add column s_nwk_s_int_key bytea,
    add column nwk_s_enc_key bytea;

update device_activation
set
    s_nwk_s_int_key = f_nwk_s_int_key,
    nwk_s_enc_key = f_nwk_s_int_key;

alter table device_activation
    alter column s_nwk_s_int_key set not null,
    alter column nwk_s_enc_key set not null;