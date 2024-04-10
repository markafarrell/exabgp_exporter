package main

import (
	"bufio"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/gizmoguy/exabgp_exporter/pkg/exporter"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
)

var (
	exaBGPCLICommand = "exabgpcli"
	exaBGPCLIRoot    = "/etc/exabgp"
)

func main() {

	var (
		_             = kingpin.Command("stream", "run in stream mode (appropriate for embedding as an exabgp process)")
		shellCmd      = kingpin.Command("standalone", "run in standalone mode (calls exabgpcli on each scrape)").Default()
		exabgpcmd     = shellCmd.Flag("exabgp.cli.command", "exabgpcli command").Default(exaBGPCLICommand).String()
		exabgproot    = shellCmd.Flag("exabgp.root", "value of --root to be passed to exabgpcli").Default(exaBGPCLIRoot).String()
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9576").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)

	promlogConfig := &promlog.Config{}

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("exabgp_exporter"))
	kingpin.HelpFlag.Short('h')
	exporterMode := kingpin.Parse()

	logger := promlog.New(promlogConfig)

	switch exporterMode {
	case "standalone":
		// nolint:errcheck
		level.Info(logger).Log(
			"msg", "Starting exabgp_exporter",
			"version", version.Info(),
			"mode", "standalone",
			"args", *exabgpcmd,
			"root", *exabgproot,
		)
		level.Info(logger).Log("buildcontext", version.BuildContext()) // nolint:errcheck
		e, err := exporter.NewStandaloneExporter(*exabgpcmd, *exabgproot, logger)
		if err != nil {
			level.Error(logger).Log("err", err) // nolint:errcheck
			os.Exit(1)
		}
		prometheus.MustRegister(e)
		prometheus.MustRegister(versioncollector.NewCollector("exabgp_exporter"))
	case "stream":
		// nolint:errcheck
		level.Info(logger).Log(
			"msg", "Starting exabgp_exporter",
			"version", version.Info(),
			"mode", "stream",
		)
		level.Info(logger).Log("buildcontext", version.BuildContext()) // nolint:errcheck
		e, err := exporter.NewEmbeddedExporter(logger)
		if err != nil {
			level.Error(logger).Log("err", err) // nolint:errcheck
			os.Exit(1)
		}
		prometheus.MustRegister(e)
		prometheus.MustRegister(versioncollector.NewCollector("exabgp_exporter"))
		reader := bufio.NewReader(os.Stdin)
		e.Run(reader)
	}
	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress) // nolint:errcheck
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
             <head><title>ExaBGP Exporter</title></head>
             <body>
             <h1>ExaBGP Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("err", err) // nolint:errcheck
		os.Exit(1)
	}
}
