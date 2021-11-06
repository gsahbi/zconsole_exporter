package probe

import (
	//	"fmt"
	"log"
	"zconsole_exporter/client"
	"zconsole_exporter/util"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

type ThreatType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func probeThreatDetails(c client.ZConsoleAPI, t Team) ([]prometheus.Metric, bool) {

	var (
		ThreatByType = prometheus.NewDesc(
			"zconsole_threat_type_count",
			"Threats type count",
			[]string{"team", "threat"}, nil,
		)
		ThreatByApp = prometheus.NewDesc(
			"zconsole_threat_byapp_count",
			"Threats breakdown by app count",
			[]string{"team", "threat", "app"}, nil,
		)
	)

	var _threattypes []ThreatType

	if err := c.Get("/api/trm/v1/threat-types", nil, &_threattypes, nil); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	threattypes := make(map[int]string)
	for _, i := range _threattypes {
		threattypes[i.Id] = i.Name
	}

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

	if err := c.Get("/api/threats/v1/stats/threats/byapp", p, nil, &json); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}
	if !gjson.Valid(json) {
		log.Print("invalid json response")
		return nil, false
	}

	m := []prometheus.Metric{}
	buckets := gjson.Get(json, `aggregations.filter\#timeRange.lterms\#threatTypeIds.buckets`)
	buckets.ForEach(func(key, value gjson.Result) bool {
		threat_id := gjson.Get(value.String(), "key").Int()
		threat_name, ok := threattypes[int(threat_id)]
		if !ok {
			threat_name = "UNKNOWN"
		}
		tcount := gjson.Get(value.String(), "doc_count").Float()
		if tcount == 0 {
			return true
		}

		m = append(m, prometheus.MustNewConstMetric(ThreatByType, prometheus.CounterValue, tcount, t.Name, threat_name))

		zapps := gjson.Get(value.String(), `sterms\#zappIds.buckets`)
		zapps.ForEach(func(key, value gjson.Result) bool {
			if count := gjson.Get(value.String(), "doc_count").Int(); count == 0 {
				return true
			}

			apps := gjson.Get(value.String(), `sterms\#zappName.buckets`)
			apps.ForEach(func(key, value gjson.Result) bool {
				app_name := gjson.Get(value.String(), "key").String()
				count := gjson.Get(value.String(), "doc_count").Float()
				m = append(m, prometheus.MustNewConstMetric(ThreatByApp, prometheus.CounterValue, count, t.Name, threat_name, app_name))
				return true // keep iterating
			})
			return true // keep iterating
		})
		return true // keep iterating
	})

	return m, true
}
