package main

import (
	"net/http"

	"github.com/alecthomas/kingpin/v2"
	"github.com/phuslu/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/tynany/cumulus_exporter/collector"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9365").String()
	telemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
)

type PhusluPromErrorLogger struct{}

func (p *PhusluPromErrorLogger) Println(v ...interface{}) {
	log.Error().Msgf("%v", v)
}

var _ promhttp.Logger = &PhusluPromErrorLogger{}

func NewPhusluPromErrorLogger() *PhusluPromErrorLogger {
	return &PhusluPromErrorLogger{}
}

func handler(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()

	_ = registry.Register(collector.NewExporter())

	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registry,
	}

	httpErrorLoggerWrapper := NewPhusluPromErrorLogger()

	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      httpErrorLoggerWrapper,
		ErrorHandling: promhttp.ContinueOnError,
	}
	promhttp.HandlerFor(gatherers, handlerOpts).ServeHTTP(w, r)
}

func parseCLI() {
	kingpin.Version(version.Print("cumulus_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {
	prometheus.MustRegister(version.NewCollector("cumulus_exporter"))

	parseCLI()

	log.Info().Msgf("Starting cumulus_exporter %s on %s", version.Info(), *listenAddress)

	http.HandleFunc(*telemetryPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Cumulus Exporter</title></head>
			<body>
			<h1>Cumulus Exporter</h1>
			<p><a href="` + *telemetryPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal().Err(err)
	}
}
