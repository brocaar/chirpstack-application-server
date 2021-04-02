alter table application
    add column payload_codec text not null default '',
    add column payload_encoder_script text not null default '',
    add column payload_decoder_script text not null default '';