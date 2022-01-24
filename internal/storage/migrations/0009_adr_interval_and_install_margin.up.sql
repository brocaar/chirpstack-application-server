alter table node
	add column adr_interval integer not null default 0,
	add column installation_margin decimal(5,2) not null default 0;