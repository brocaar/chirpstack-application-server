alter table application
	drop column rx_delay,
	drop column rx1_dr_offset,
	drop column channel_list_id,
	drop column rx_window,
	drop column rx2_dr,
	drop column relax_fcnt,
	drop column adr_interval,
	drop column installation_margin,
	drop column is_abp,
	drop column is_class_c;

alter table node
	drop column use_application_settings;