alter table device_profile
    drop column payload_codec,
    drop column payload_encoder_script,
    drop column payload_decoder_script;