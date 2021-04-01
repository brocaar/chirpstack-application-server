alter table device
    drop column device_status_external_power_source,
    alter column device_status_battery type integer;