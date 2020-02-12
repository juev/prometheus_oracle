# Prometheus Oracle Exporter

# Description

A [Prometheus](https://prometheus.io/) exporter for Oracle. All requests are launched in parallel processes and do not block the HTTP entry point of Prometheus.

# Installation

## Docker

From the very beginning, you need to create a configuration file with a list of all connections and all the requests that you need, an example file in `config.yaml.template`.
Just copy to `config.yaml` and fill it.

You can run via Docker using an existing image. If you don't already have an Oracle server, you can run one locally in a container and then link the exporter to it.

```bash
docker run -d --name oracle -p 1521:1521 wnameless/oracle-xe-11g:16.04
docker run -d -v /Users/username/prometheus_oracle/config.yaml:/config.yaml -v /Users/username/prometheus_oracle/log.json:/log.json --name prometheus_oracle --link=oracle -p 9101:9101 juev/prometheus_oracle
```

Or just:

```bash
docker run --rm -v /Users/username/prometheus_oracle/config.yaml:/config.yaml -v /Users/username/prometheus_oracle/log.json:/log.json --name prometheus_oracle -p 9101:9101 juev/prometheus_oracle
```

# Metrics

The following metrics are exposed currently.

- oracledb_exporter_db_metric
- oracledb_exporter_up

Example:

```
# HELP oracledb_exporter_db_metric Business metrics from Database
# TYPE oracledb_exporter_db_metric gauge
oracledb_exporter_db_metric{database="DB1",name="name1",value="101598180.65"} 1.0159818065e+08
oracledb_exporter_db_metric{database="DB1",name="name2",value="Text1"} 0
oracledb_exporter_db_metric{database="DB2",name="name1",value=""} 1
oracledb_exporter_db_metric{database="DB2",name="name2",value=""} 1
oracledb_exporter_db_metric{database="DB2",name="name3",value="Text2"} 0
# HELP oracledb_exporter_up Database status
# TYPE oracledb_exporter_up gauge
oracledb_exporter_up{database="DB1"} 1
oracledb_exporter_up{database="DB2"} 1
```
