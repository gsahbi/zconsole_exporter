package probe

import (
	"fmt"
	"log"
	"zconsole_exporter/client"
	"zconsole_exporter/util"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

func probeThreatStats(c client.ZConsoleAPI, t Team) ([]prometheus.Metric, bool) {

	var (
		Thrts = prometheus.NewDesc(
			"zconsole_threat_count",
			"Threats vectors count",
			[]string{"team", "vector", "class"}, nil,
		)
	)

	vectors := map[string]string{"1": "network", "2": "device", "3": "malware"}
	severities := map[string]string{"2": "critical", "3": "elevated"}

	after, before, err := util.GetTodayRange()
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	var json string

	p := Options{}
	p.After = after
	p.Before = before
	p.TeamId = t.Id
	if err := c.Get("api/threats/v1/stats/vectors", p, nil, &json); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	if !gjson.Valid(json) {
		log.Print("invalid json response")
		return nil, false
	}

	m := []prometheus.Metric{}
	res := gjson.Get(json, `aggregations.filter*.filter*.buckets`)

	for key, vect := range vectors {
		v := gjson.Get(res.String(), fmt.Sprintf("%s.filters\\#severities.buckets", key))

		for k, sevr := range severities {
			i := gjson.Get(v.String(), fmt.Sprintf("%s.doc_count", k)).Float()
			m = append(m, prometheus.MustNewConstMetric(Thrts, prometheus.CounterValue, i, t.Name, vect, sevr))
		}
		mtg := gjson.Get(res.String(), fmt.Sprintf("%s.filters*.buckets.0.doc_count", key)).Float()
		tot := gjson.Get(res.String(), fmt.Sprintf("%s.doc_count", key)).Float()
		m = append(m, prometheus.MustNewConstMetric(Thrts, prometheus.CounterValue, mtg, t.Name, vect, "mitigated"))
		m = append(m, prometheus.MustNewConstMetric(Thrts, prometheus.CounterValue, tot, t.Name, vect, "total"))

	}

	return m, true
}
