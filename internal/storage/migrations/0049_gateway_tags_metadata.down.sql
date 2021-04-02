drop index idx_gateway_tags;

alter table gateway
    drop column metadata,
    drop column tags;