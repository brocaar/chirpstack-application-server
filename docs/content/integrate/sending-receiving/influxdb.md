---
title: InfluxDB
menu:
    main:
        parent: sending-receiving
---

# InfluxDB integration

When configured, the InfluxDB integration will write device data into an
[InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/) database.
This makes it possible to directly visualize all device data using for example
[Grafana](https://grafana.com).

* [Getting started with InfluxDB](https://docs.influxdata.com/influxdb/latest/)
* [Getting started with Grafana](http://docs.grafana.org)

## Requirements

Before this integration is able to write data into InfluxDB, the uplink
payloads must be decoded. The payload codec can be configured per
[device-profile]({{<ref "use/device-profiles.md">}}). To validate that the uplink
payloads are decoded, you can use the [live device event-log]({{<ref "use/event-logging.md">}})
feature. Decoded payload data will be available under the `object` key in
the JSON object.

## Measurements

### Naming

All measurements are using the field names from the `object` element,
joined by an underscore (`_`) in case the object element is nested.
Payload data is prefixed by `device_frmpayload_data`.

Example:

{{<highlight json>}}
{
    "object": {
        "temperature_sensor": {
            "1": 23.5
        }
    }
}
{{< /highlight >}}

The above will translate to the measurement `device_frmpayload_data_temperature_sensor_1`.

**Note:** When using the [CayenneLPP codec]({{<ref "use/device-profiles.md">}})
`camelCasing` is used when the data is presented as JSON. However, for the InfluxDB
naming, `snake_casing` is used. Thus `temperatureSensor` in JSON translates to
`temperature_sensor` as measurement name in InfluxDB.

### Location data

When both `latitude` and `longitude` keys are found (on the same level within
the `object`), both measurements are treated as a single geolocation measurement.

Example:

{{<highlight json>}}
{
    "object": {
        "latitude": 1.123,
        "longitude": 2.123
    }
}
{{< /highlight >}}

The above will translate to the measurement `device_frmpayload_data_location`
with values `latitude`, `longitude` and `geohash` (see also [Geohash](https://en.wikipedia.org/wiki/Geohash)).

## Tags

For aggregation, each measurement will have the following tags (additional to
the device tags):

* `application_name`
* `device_name`
* `dev_eui`
* `f_port` (LoRaWAN port used for uplink)

## Device uplink meta-data

For analyzing and monitoring the usage of spreading-factors, channels, etc.
the InfluxDB integration will also write a measurement named `device_uplink`
with as values a counter `value` 1, `rssi`, `snr` and `f_cnt`. For Aggregation,
the following tags are available (additional to the device tags):

* `application_name`
* `device_name`
* `dev_eui`
* `dr`
* `frequency`

## Device battery status (deprecated)

When this information is available, the device battery status will be written
to the measurement name `device_status_battery`. For aggregation, the following
tags are available (additional to the device tags):

* `application_name`
* `device_name`
* `dev_eui`

## Device battery level

When this information is available, the device battery level (percentage)
will be written to the measurement name `device_status_battery_level`. For
aggregation, the following tags are available (additional to the device tags):

* `application_name`
* `device_name`
* `dev_eui`

## Device margin status

When this information is available, the device margin status will be written
to the measurement name `device_status_margin`. For aggregation, the following
tags are available (additional to the device tags):

* `application_name`
* `device_name`
* `dev_eui`
