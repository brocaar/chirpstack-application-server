create index idx_node_search on node using gin (name gin_trgm_ops, encode(dev_eui, 'hex') gin_trgm_ops);