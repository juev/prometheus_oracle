package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/kkyr/fig"
	"github.com/mattn/go-oci8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

type Configuration struct {
	Host         string `fig:"host,default=0.0.0.0"`
	Port         string `fig:"port,default=9101"`
	QueryTimeout string `fig:"querytimeout,default=30"`
	Databases    []Database
}

type Database struct {
	Dsn          string
	Host         string  `fig:",default=127.0.0.1"`
	User         string  `fig:"user"`
	Password     string  `fig:"password"`
	Database     string  `fig:"database"`
	Port         string  `fig:"port,default=1522"`
	MaxIdleConns string  `fig:",default=10"`
	MaxOpenConns string  `fig:",default=10"`
	Queries      []Query `fig:"queries"`
	db           *sql.DB
}

type Query struct {
	Sql      string `fig:"sql"`
	Name     string `fig:"name"`
	Interval string `fig:"interval,default=1"`
}

const (
	namespace = "oracledb"
	exporter  = "exporter"
)

var (
	configuration Configuration
	metricMap     map[string]*prometheus.GaugeVec
	timeout       int
	maxIdleConns  int
	maxOpenConns  int
	err           error
	configFile    string
	logFile       string
)

func init() {
	flag.StringVarP(&configFile, "configFile", "c", "config.yaml", "Config file name (default: config.yaml)")
	flag.StringVarP(&logFile, "logFile", "l", "stdout", "Log filename (default: stdout)")

	metricMap = map[string]*prometheus.GaugeVec{
		"value": prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "query_value",
			Help:      "Value of Business metrics from Database",
		}, []string{"database", "name"}),
		"error": prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "query_error",
			Help:      "Result of last query, 1 if we have errors on running query",
		}, []string{"database", "name"}),
		"duration": prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "query_duration_seconds",
			Help:      "Duration of the query in seconds",
		}, []string{"database", "name"}),
		"up": prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "up",
			Help:      "Database status",
		}, []string{"database"}),
	}
	for _, metric := range metricMap {
		prometheus.MustRegister(metric)
	}
}

func execQuery(database Database, query Query) {

	defer func(begun time.Time) {
		duration := time.Since(begun).Seconds()
		metricMap["duration"].WithLabelValues(database.Database, query.Name).Set(duration)
	}(time.Now())

	// Reconnect if we lost connection
	if err := database.db.Ping(); err != nil {
		if strings.Contains(err.Error(), "sql: database is closed") {
			logrus.Infoln("Reconnecting to DB: ", database.Database)
			database.db, _ = sql.Open("oci8", database.Dsn)
			database.db.SetMaxIdleConns(maxIdleConns)
			database.db.SetMaxOpenConns(maxOpenConns)
		}
	}

	// Validate connection
	if err := database.db.Ping(); err != nil {
		logrus.Errorf("Error on connect to database '%s': %v", database.Database, err)
		metricMap["up"].WithLabelValues(database.Database).Set(0)
		return
	} else {
		metricMap["up"].WithLabelValues(database.Database).Set(1)
	}

	// query db
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	rows, err := database.db.QueryContext(ctx, query.Sql)
	if ctx.Err() == context.DeadlineExceeded {
		logrus.Errorf("oracle query '%s' timed out", query.Name)
		metricMap["error"].WithLabelValues(database.Database, query.Name).Set(1)
		return
	}
	if err != nil {
		logrus.Errorf("oracle query '%s' failed: %v", query.Name, err)
		metricMap["error"].WithLabelValues(database.Database, query.Name).Set(1)
		return
	}

	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	defer func() {
		err := rows.Close()
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	for rows.Next() {
		for i := range cols {
			vals[i] = &vals[i]
		}

		err = rows.Scan(vals...)
		if err != nil {
			break
		}

		for i := range cols {
			metricMap["error"].WithLabelValues(database.Database, query.Name).Set(0)
			if vals[i] == nil {
				metricMap["value"].WithLabelValues(database.Database, query.Name).Set(0)
			} else {
				val, err := strconv.ParseFloat(strings.TrimSpace(vals[i].(string)), 64)
				if err != nil {
					logrus.Errorf("Cannot convert value '%s' to float on query '%s': %v", vals[i].(string), query.Name, err)
					metricMap["error"].WithLabelValues(database.Database, query.Name).Set(1)
					return
				}
				metricMap["value"].WithLabelValues(database.Database, query.Name).Set(val)
			}
		}
	}
}

func main() {
	flag.Parse()
	if logFile == "stdout" {
		logrus.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.Info("Failed to log to file, using default stdout")
			logrus.SetOutput(os.Stdout)
		}
	}

	err = fig.Load(&configuration, fig.File(configFile))
	if err != nil {
		logrus.Fatal("Fatal error on reading configuration: ", err)
	}

	timeout, err = strconv.Atoi(configuration.QueryTimeout)
	if err != nil {
		logrus.Fatal("error while converting timeout option value: ", err)
	}

	for _, database := range configuration.Databases {
		// connect to database
		database.Dsn = oci8.QueryEscape(database.User) + "/" + oci8.QueryEscape(database.Password) +
			"@" + database.Host + ":" + database.Port + "/" + database.Database
		logrus.Infoln("Connecting to DB: ", database.Database)
		database.db, err = sql.Open("oci8", database.Dsn)
		if err != nil {
			logrus.Errorln("Error connecting to db: ", err)
		}
		maxIdleConns, err = strconv.Atoi(database.MaxIdleConns)
		if err != nil {
			logrus.Fatal("error while converting maxIdleConns option value: ", err)
		}

		maxOpenConns, err = strconv.Atoi(database.MaxOpenConns)
		if err != nil {
			logrus.Fatal("error while converting maxOpenConns option value: ", err)
		}

		database.db.SetMaxIdleConns(maxOpenConns)
		database.db.SetMaxOpenConns(maxOpenConns)

		// create cron jobs for every query on database
		if err := database.db.Ping(); err == nil {
			for _, query := range database.Queries {
				gocron.Every(5).Minutes().DoSafely(execQuery, database, query)
			}
		} else {
			logrus.Errorf("Error connecting to db '%s': %v", database.Database, err)
		}
	}

	gocron.Start()
	gocron.RunAll()

	prometheusConnection := configuration.Host + ":" + configuration.Port
	logrus.Printf("listen: %s", prometheusConnection)
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(prometheusConnection, nil)
	if err != nil {
		logrus.Fatalln("Fatal error on serving metrics:", err)
	}
}
