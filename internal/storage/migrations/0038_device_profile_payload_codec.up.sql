alter table device_profile
    add column payload_codec text not null default '',
    add column payload_encoder_script text not null default '',
    add column payload_decoder_script text not null default '';

alter table device_profile
    alter column payload_codec drop default,
    alter column payload_encoder_script drop default,
    alter column payload_decoder_script drop default;