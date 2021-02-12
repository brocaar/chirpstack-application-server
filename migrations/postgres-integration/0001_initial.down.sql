DROP INDEX IF EXISTS idx_device_up_received_at;
DROP INDEX IF EXISTS idx_device_up_dev_eui;
DROP INDEX IF EXISTS idx_device_up_application_id;
DROP INDEX IF EXISTS idx_device_up_frequency;
DROP INDEX IF EXISTS idx_device_up_dr;
DROP INDEX IF EXISTS idx_device_up_tags;
DROP TABLE IF EXISTS device_up;

DROP INDEX IF EXISTS idx_device_status_received_at;
DROP INDEX IF EXISTS idx_device_status_dev_eui;
DROP INDEX IF EXISTS idx_device_status_application_id;
DROP INDEX IF EXISTS idx_device_status_tags;
DROP TABLE IF EXISTS device_status;

DROP INDEX IF EXISTS idx_device_join_received_at;
DROP INDEX IF EXISTS idx_device_join_dev_eui;
DROP INDEX IF EXISTS idx_device_join_application_id;
DROP INDEX IF EXISTS idx_device_join_tags;
DROP TABLE IF EXISTS device_join;

DROP INDEX IF EXISTS idx_device_ack_received_at;
DROP INDEX IF EXISTS idx_device_ack_dev_eui;
DROP INDEX IF EXISTS idx_device_ack_application_id;
DROP INDEX IF EXISTS idx_device_ack_tags;
DROP TABLE IF EXISTS device_ack;

DROP INDEX IF EXISTS idx_device_error_received_at;
DROP INDEX IF EXISTS idx_device_error_dev_eui;
DROP INDEX IF EXISTS idx_device_error_application_id;
DROP INDEX IF EXISTS idx_device_error_tags;
DROP TABLE IF EXISTS device_error;

DROP INDEX IF EXISTS idx_device_location_received_at;
DROP INDEX IF EXISTS idx_device_location_dev_eui;
DROP INDEX IF EXISTS idx_device_location_application_id;
DROP INDEX IF EXISTS idx_device_location_tags;
DROP TABLE IF EXISTS device_location;