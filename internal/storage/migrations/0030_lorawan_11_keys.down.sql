alter table device_activation
    drop column nwk_s_enc_key,
    drop column s_nwk_s_int_key;

alter table device_activation
    rename column f_nwk_s_int_key to nwk_s_key;

alter table device_keys
    drop column app_key;

alter table device_keys
    rename column nwk_key to app_key;