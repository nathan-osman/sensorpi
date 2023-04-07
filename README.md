## sensorpi

This handy little program is designed to monitor sensors connected to a Raspberry Pi and react to the data based on a configuration file. For example, a configuration file for reading temperature data from a [1-Wire](https://en.wikipedia.org/wiki/1-Wire) sensor connected to GPIO 4 and sending the data to [InfluxDB](https://en.wikipedia.org/wiki/InfluxDB) every five minutes would look like this:

```yaml
input:
  onewire:
    - name: 28-0516a43c9fff
      interval: 5m
      output:
        influxdb:
          name: temperature
          tags:
            location: garage
output:
  influxdb:
    addr: 127.0.0.1
    username: username
    password: password
    database: example
```
