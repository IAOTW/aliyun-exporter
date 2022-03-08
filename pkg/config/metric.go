package config

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Metric meta
type Metric struct {
	Name        string   `json:"name"`
	Alias       string   `json:"alias,omitempty"`
	Period      string   `json:"period,omitempty"`
	Description string   `json:"desc,omitempty"`
	Dimensions  []string `json:"dimensions,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Measure     string   `json:"measure,omitempty"`
	Format      bool     `json:"format,omitempty"`
	desc        *prometheus.Desc
}

// setdefault options
func (m *Metric) setDefaults() {
	if m.Period == "" {
		m.Period = "60"
	}
	if m.Description == "" {
		m.Description = m.Name
	}
	// Do some fallback in case someone runs this exporter directly
	// without modifying the example configuration
	periods := strings.Split(m.Period, ",")
	m.Period = periods[0]
	measures := strings.Split(m.Measure, ",")
	m.Measure = measures[0]
	switch m.Measure {
	case "Maximum", "Minimum", "Average", "Value":
	default:
		m.Measure = "Average"
	}
	m.Description = fmt.Sprintf("%s unit:%s measure:%s", m.Description, m.Unit, m.Measure)
}

// String name of metric
func (m *Metric) String() string {
	if m.Alias != "" {
		return m.Alias
	}
	m.Name = strings.Replace(m.Name, ".", "_", -1)
	if m.Format {
		return strings.Join([]string{m.Name, formatUnit(m.Unit)}, "_")
	}
	return m.Name
}

// Desc to prometheus desc
func (m *Metric) Desc(ns, sub string, dimensions ...string) *prometheus.Desc {
	if len(m.Dimensions) == 0 {
		m.Dimensions = dimensions
	}
	if m.desc == nil {
		m.desc = prometheus.NewDesc(
			strings.Join([]string{ns, sub, m.String()}, "_"),
			m.Description,
			append(m.Dimensions, "cloudID"),
			nil,
		)
	}
	return m.desc
}

var durationUnitMapping = map[string]string{
	"s": "second",
	"m": "minute",
	"h": "hour",
	"d": "day",
}

func formatUnit(s string) string {
	s = strings.TrimSpace(s)
	if s == "%" {
		return "percent"
	}

	if strings.IndexAny(s, "/") > 0 {
		fields := strings.Split(s, "/")
		if len(fields) == 2 {
			if v, ok := durationUnitMapping[fields[1]]; ok {
				return strings.ToLower(strings.Join([]string{fields[0], "per", v}, "_"))
			}
		}
	}
	return strings.ToLower(s)
}
