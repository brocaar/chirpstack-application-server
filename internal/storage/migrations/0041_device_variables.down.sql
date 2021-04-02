drop index idx_device_tags;

alter table device
    drop column variables,
    drop column tags;