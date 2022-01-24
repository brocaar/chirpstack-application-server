alter table node
	add column rx_window int2 not null default 0,
	add column rx2_dr int2 not null default 0;
