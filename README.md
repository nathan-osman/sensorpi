![sensorpi logo](https://github.com/nathan-osman/sensorpi/blob/main/img/logo.png?raw=true)

[![GoDoc](https://godoc.org/github.com/nathan-osman/sensorpi?status.svg)](https://godoc.org/github.com/nathan-osman/sensorpi)
[![MIT License](https://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](https://opensource.org/licenses/MIT)

This handy little program is designed to monitor sensors connected to a Raspberry Pi and react to the data based on a configuration file.

For example, a configuration file for reading temperature data from a [1-Wire](https://en.wikipedia.org/wiki/1-Wire) sensor connected to GPIO 4 and sending the data to [InfluxDB](https://en.wikipedia.org/wiki/InfluxDB) every five minutes would look like this:

```yaml
plugins:
  influxdb:
    url: "http://127.0.0.1:8086"
    username: username
    password: password
    database: example
connections:
  - input:
      plugin: onewire
      parameters:
        device: 28-0516a43c9fff
    outputs:
      - plugin: influxdb
        parameters:
          name: temperature
          tags:
            location: garage
    interval: 5m
```
