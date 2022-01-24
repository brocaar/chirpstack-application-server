alter table device_activation
    add column f_nwk_s_int_key bytea,
    add column s_nwk_s_int_key bytea,
    add column nwk_s_enc_key bytea;