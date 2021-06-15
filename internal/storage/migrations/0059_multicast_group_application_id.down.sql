alter table multicast_group
    add column service_profile_id uuid null references service_profile;

create index idx_multicast_group_service_profile_id on multicast_group (service_profile_id);

update multicast_group
    set
        service_profile_id = a.service_profile_id
    from (
        select
            service_profile_id,
            id
        from
            application
    ) as a
    where
        multicast_group.application_id = a.id;


alter table multicast_group
    drop column application_id,
    alter column service_profile_id set not null;
