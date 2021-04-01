alter table application
	add column rx_delay int2 not null default 0,
	add column rx1_dr_offset int2 not null default 0,
	add column channel_list_id bigint references channel_list on delete set null,
	add column rx_window int2 not null default 0,
	add column rx2_dr int2 not null default 0,
	add column relax_fcnt boolean not null default false,
	add column adr_interval integer not null default 0,
	add column installation_margin decimal(5,2) not null default 0,
	add column is_abp boolean not null default false,
	add column is_class_c boolean not null default false;

alter table node
	add column use_application_settings boolean not null default false;
