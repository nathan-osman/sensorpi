![sensorpi logo](https://github.com/nathan-osman/sensorpi/blob/main/img/logo.png?raw=true)

[![GoDoc](https://godoc.org/github.com/nathan-osman/sensorpi?status.svg)](https://godoc.org/github.com/nathan-osman/sensorpi)
[![MIT License](https://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](https://opensource.org/licenses/MIT)

This handy little program is designed to monitor sensors connected to a Raspberry Pi and react to the data based on a configuration file.

### Overview

There are several different types of plugins available:

- **trigger plugins** wait for something to happen (motion detected, etc.)
- **input plugins** read a value at a regular interval (such as temperature, etc.)
- **output plugins** do something with a value (write to database, etc.)

In order to do something useful, input and trigger plugins need to be connected to output plugins. For example, you might connect an input plugin for reading temperature values to an output plugin that sends the values to a database.

### Plugin List

This is an exhaustive list of plugins available and a brief description of how they can be used.

| Name           | Type            | Description                  |
| -------------- | --------------- | ---------------------------- |
| bme280         | input           | read from a BME280 sensor    |
| command        | output          | run a command                |
| console        | output          | output to the console        |
| daylight       | input, trigger  | sunrise / sunset times       |
| gpio           | output, trigger | GPIO I/O                     |
| grove-moisture | input           | read moisture values         |
| influxdb       | output          | write to InfluxDB            |
| mqtt           | output, trigger | watch, publish MQTT topic    |
| nut            | input           | read values from NUT server  |
| onewire        | input           | read from 1-Wire sensor      |
| timer          | trigger         | trigger at regular intervals |

### Example

A simple configuration file for reading temperature data from a [1-Wire](https://en.wikipedia.org/wiki/1-Wire) sensor connected to GPIO 4 and sending the data to [InfluxDB](https://en.wikipedia.org/wiki/InfluxDB) every five minutes would look like this:

```yaml
plugins:
  influxdb:
    url: "http://127.0.0.1:8086"
    username: username
    password: password
    database: example
inputs:
  - plugin: onewire
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
