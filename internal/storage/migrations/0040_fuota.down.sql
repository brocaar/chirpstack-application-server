drop table fuota_deployment_device;

drop index idx_fuota_deployment_next_step_after;
drop index idx_fuota_deployment_state;
drop index idx_fuota_deployment_multicast_group_id;
drop table fuota_deployment;

drop index idx_remote_fragmentation_session_retry_after;
drop index idx_remote_fragmentation_session_state_provisioned;
drop table remote_fragmentation_session;

drop index idx_remote_multicast_class_c_session_state_retry_after;
drop index idx_remote_multicast_class_c_session_state_provisioned;
drop table remote_multicast_class_c_session;

drop index idx_remote_multicast_setup_retry_after;
drop index idx_remote_multicast_setup_state_provisioned;
drop table remote_multicast_setup;

alter table device_keys
    drop column gen_app_key;

alter table multicast_group
    drop column mc_key,
    drop column f_cnt;