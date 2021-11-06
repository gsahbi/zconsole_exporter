package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"zconsole_exporter/client"
	"zconsole_exporter/util/config"

	"github.com/prometheus/client_golang/prometheus"
)



type Options struct {
	After  string `url:"after"`
	Before string `url:"before"`
	TeamId string `url:"teamId.keyword"`
}

type Results struct {
	Aggs map[string]interface{} `json:"aggregations"`
}

type ProbeCollector struct {
	metrics []prometheus.Metric
}

type probeFunc func(client.ZConsoleAPI, Team) ([]prometheus.Metric, bool)

func (p *ProbeCollector) Probe(ctx context.Context, target string, hc *http.Client, savedConfig config.Config) (bool, error) {
	tgt, err := url.Parse(target)
	if err != nil {
		return false, fmt.Errorf("url.Parse failed: %v", err)
	}

	if tgt.Scheme != "https" && tgt.Scheme != "http" {
		return false, fmt.Errorf("unsupported scheme %q", tgt.Scheme)
	}

	// Filter anything else than scheme and hostname
	u := url.URL{
		Scheme: tgt.Scheme,
		Host:   tgt.Host,
	}
	c, err := client.ZConsoleRequest(ctx, u, hc, savedConfig)
	if err != nil {
		return false, err
	}

	// Probe team list
	ts, er := probeTeams(c)
	if er != nil {
		return false, er
	}


	// TODO: Make parallel
	success := true
	for _, team := range ts {
		for _, f := range []probeFunc{
			probeThreatStats,
			probeThreatDetails,
		} {
			m, ok := f(c, team)
			if !ok {
				success = false
			}
			p.metrics = append(p.metrics, m...)
		}		
	}


	return success, nil
}

func (p *ProbeCollector) Collect(c chan<- prometheus.Metric) {
	// Collect result of new probe functions
	for _, m := range p.metrics {
		c <- m
	}
}

func (p *ProbeCollector) Describe(c chan<- *prometheus.Desc) {
}
