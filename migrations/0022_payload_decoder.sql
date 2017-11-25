-- +migrate Up
alter table application
    add column payload_codec text not null default '',
    add column payload_encoder_script text not null default '',
    add column payload_decoder_script text not null default '';

-- +migrate Down
alter table  application
    drop column payload_codec,
    drop column payload_encoder_script,
    drop column payload_decoder_script;
