alter table multicast_group
    add column application_id bigint null references application on delete cascade;

create index idx_multicast_group_application_id on multicast_group (application_id);

update multicast_group
    set
        application_id = meta.application_id
    from (
        select
            d.application_id,
            dmg.multicast_group_id
        from device d
        inner join device_multicast_group dmg
            on dmg.dev_eui = d.dev_eui
    ) as meta
    where
        multicast_group.id = meta.multicast_group_id;

alter table multicast_group
    drop column service_profile_id,
    alter column application_id set not null;
