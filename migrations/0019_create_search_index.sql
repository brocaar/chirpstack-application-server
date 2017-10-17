-- +migrate Up
create index idx_node_dev_eui_prefix on node (encode(dev_eui, 'hex') varchar_pattern_ops);
create index idx_node_name_prefix on node (name varchar_pattern_ops);

-- +migrate Down
drop index idx_node_dev_eui_prefix;
drop index idx_node_name_prefix;
