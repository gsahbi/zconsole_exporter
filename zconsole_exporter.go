package main

import (
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"zconsole_exporter/probe"
	"zconsole_exporter/util/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version = "(devel)"
	GitHash = "(no hash)"
)

type BuildInfo struct {
	version   string
	gitHash   string
	goVersion string
}

func setUpMetricsEndpoint(buildInfo BuildInfo) {
	zConsoleExporterInfo := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "zconsole_exporter_build_info",
		Help: "This info metric contains build information for about the exporter",
	}, []string{"version", "revision", "goversion"})

	zConsoleExporterInfo.With(prometheus.Labels{
		"version":   buildInfo.version,
		"revision":  buildInfo.gitHash,
		"goversion": buildInfo.goVersion,
	}).Set(1)
}

func getBuildInfo() BuildInfo {
	// don't overwrite the version if it was set by -ldflags=-X
	if info, ok := debug.ReadBuildInfo(); ok && Version == "(devel)" {
		mod := &info.Main
		if mod.Replace != nil {
			mod = mod.Replace
		}
		Version = mod.Version
	}
	// remove leading `v`
	massagedVersion := strings.TrimPrefix(Version, "v")
	buildInfo := BuildInfo{
		version:   massagedVersion,
		gitHash:   GitHash,
		goVersion: runtime.Version(),
	}
	return buildInfo
}

func main() {
	buildInfo := getBuildInfo()
	log.Printf("zConsoleExporter %s ( %s )", buildInfo.version, buildInfo.gitHash)
	setUpMetricsEndpoint(buildInfo)


	if err := config.Init(); err != nil {
		log.Fatalf("Initialization error: %+v", err)
	}

	savedConfig := config.GetConfig()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probe.ProbeHandler)
	go func() {
		if err := http.ListenAndServe(savedConfig.Listen, nil); err != nil {
			log.Fatalf("Unable to serve: %v", err)
		}
	}()
	log.Printf("zConsole exporter running, listening on %q", savedConfig.Listen)
	select {}
}