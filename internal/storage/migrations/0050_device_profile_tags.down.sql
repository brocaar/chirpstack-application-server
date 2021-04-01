drop index idx_device_profile_tags;

alter table device_profile
    drop column tags;