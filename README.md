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
docker run -d -v /Users/username/prometheus_oracle/config.yaml:/config.yaml --name prometheus_oracle --link=oracle -p 9101:9101 juev/prometheus_oracle
```

Or just:

```bash
docker run --rm -v /Users/username/prometheus_oracle/config.yaml:/config.yaml --name prometheus_oracle -p 9101:9101 juev/prometheus_oracle
```

# Metrics

The following metrics are exposed currently.

- oracledb_exporter_dbmetric
- oracledb_exporter_query_duration_seconds
- oracledb_exporter_query_error
- oracledb_exporter_up

Example:

```
# HELP oracledb_exporter_dbmetric Value of Business metrics from Database
# TYPE oracledb_exporter_dbmetric gauge
oracledb_exporter_dbmetric{database="DB1",name="metric1"} 1.59862528e+06
oracledb_exporter_dbmetric{database="DB1",name="metric2"} 0
# HELP oracledb_exporter_query_duration Duration of the query in seconds
# TYPE oracledb_exporter_query_duration gauge
oracledb_exporter_query_duration_seconds{database="DB1",name="metric1"} 0.005608735
oracledb_exporter_query_duration_seconds{database="DB1",name="metric2"} 0.008431937
oracledb_exporter_query_duration_seconds{database="DB2",name="metric3"} 0.031262699
oracledb_exporter_query_duration_seconds{database="DB2",name="metric4"} 0.01056187
oracledb_exporter_query_duration_seconds{database="DB2",name="metric5"} 0.006411418
oracledb_exporter_query_duration_seconds{database="DB2",name="metric6"} 0.00876666
oracledb_exporter_query_duration_seconds{database="DB2",name="metric7"} 0.027751112
oracledb_exporter_query_duration_seconds{database="DB2",name="metric8"} 0.008800151
oracledb_exporter_query_duration_seconds{database="DB2",name="metric9"} 0.012350262
# HELP oracledb_exporter_query_error Result of last query, 1 if we have errors on running query
# TYPE oracledb_exporter_query_error gauge
oracledb_exporter_query_error{database="DB1",name="metric1"} 0
oracledb_exporter_query_error{database="DB1",name="metric2"} 0
oracledb_exporter_query_error{database="DB2",name="metric3"} 0
oracledb_exporter_query_error{database="DB2",name="metric4"} 0
oracledb_exporter_query_error{database="DB2",name="metric5"} 0
oracledb_exporter_query_error{database="DB2",name="metric6"} 0
oracledb_exporter_query_error{database="DB2",name="metric7"} 0
oracledb_exporter_query_error{database="DB2",name="metric8"} 0
oracledb_exporter_query_error{database="DB2",name="metric9"} 0
# HELP oracledb_exporter_up Database status
# TYPE oracledb_exporter_up gauge
oracledb_exporter_up{database="DB1"} 1
oracledb_exporter_up{database="DB2"} 1
```

# Config file

```
# Host, default value `0.0.0.0` (optional)
host: 0.0.0.0
# Port, default value `9101` (optional)
port: 9101
# QueryTimeout, default value `30` in seconds (optional)
queryTimeout: 5

# Array of databases and queries
databases:
    # Host, default value `127.0.0.1`, hostname for DB (optional)
  - host: 'dummy'
    # User (required)
    user: dummy
    # Port, default value `1522` (optional)
    port: 1522
    # Password (required)
    password: 'password'
    # Database name (required)
    database: dummy
    # MaxIdleConns, default value `10` (optional)
    maxIdleConns: 10
    # MaxOpenConns, default value `10` (optional)
    maxOpenConns: 10
    # Aray of queries
    queries:
        # SQL, query (required)
      - sql: "select numbers1 from dummy"
        # SQL name (required)
        name: value1
        # Interval between queries, default value `1` in minites (optional)
        interval: 1
```