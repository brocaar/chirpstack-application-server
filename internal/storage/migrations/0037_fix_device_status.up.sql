alter table device
    add column device_status_external_power_source boolean not null default false,
    alter column device_status_battery type decimal(5,2);

alter table device
    alter column device_status_external_power_source drop default;