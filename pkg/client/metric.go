package client

import (
	"encoding/json"
	"fmt"
	"github.com/IAOTW/aliyun-exporter/pkg/config"
	cms "github.com/alibabacloud-go/cms-20190101/v7/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"sort"
)

var ignores = map[string]struct{}{
	"timestamp": {},
	"Maximum":   {},
	"Minimum":   {},
	"Average":   {},
}

// map for all avaliable namespaces
// todo: is there a way to add desc into yaml file?
var allNamespaces = map[string]string{
	"acs_ecs_dashboard":              "Elastic Compute Service (ECS)",
	"acs_containerservice_dashboard": "Container Service for Swarm",
	"acs_kubernetes":                 "Container Service for Kubernetes (ACK)",
	"acs_oss_dashboard":              "Object Storage Service (OSS)",
	"acs_slb_dashboard":              "Server Load Balancer (SLB)",
	"acs_vpc_eip":                    "Elastic IP addresses (EIPs)",
	"acs_nat_gateway":                "NAT Gateway",
	"acs_anycast_eip":                "Anycast Elastic IP address (EIP)",
	"acs_rds_dashboard":              "ApsaraDB RDS",
	"acs_mongodb":                    "ApsaraDB for MongoDB",
	"acs_memcache":                   "ApsaraDB for Memcache",
	"acs_kvstore":                    "ApsaraDB for Redis",
	"acs_hitsdb":                     "Time Series Database (TSDB)",
	"acs_clickhouse":                 "ClickHouse",
	"acs_cds":                        "ApsaraDB for Cassandra",
	"waf":                            "Web Application Firewall (WAF)",
	"acs_elasticsearch":              "Elasticsearch",
	"acs_mns_new":                    "queues of Message Service (MNS)",
	"acs_kafka":                      "Message Queue for Apache Kafka",
	"acs_amqp":                       "Alibaba Cloud Message Queue for AMQP instances",
}

// AllNamespaces return allNamespaces
func AllNamespaces() map[string]string {
	return allNamespaces
}

// allNamesOfNamespaces return all avaliable namespaces
func allNamesOfNamespaces() []string {
	ss := make([]string, 0, len(allNamespaces))
	for name := range allNamespaces {
		ss = append(ss, name)
	}
	return ss
}

// Datapoint datapoint
type Datapoint map[string]interface{}

// Get return value for measure
func (d Datapoint) Get(measure string) float64 {
	v, ok := d[measure]
	if !ok {
		return 0
	}
	return v.(float64)
}

// Labels return labels that not in ignores
func (d Datapoint) Labels() []string {
	labels := make([]string, 0)
	for k := range d {
		if _, ok := ignores[k]; !ok {
			labels = append(labels, k)
		}
	}
	sort.Strings(labels)
	return labels
}

// Values return values for lables
func (d Datapoint) Values(labels ...string) []string {
	values := make([]string, 0, len(labels))
	for i := range labels {
		values = append(values, fmt.Sprintf("%s", d[labels[i]]))
	}
	return values
}

// MetricClient wrap cms.client
type MetricClient struct {
	cloudID string
	cms     *cms.Client
	logger  log.Logger
}

// NewMetricClient create metric Client
func NewMetricClient(cloudID, ak, secret, region string, logger log.Logger) (*MetricClient, error) {
	c := &openapi.Config{
		AccessKeyId:     &ak,
		AccessKeySecret: &secret,
		SecurityToken:   &region,
	}
	c = c.SetReadTimeout(50)
	cmsClient := &cms.Client{}
	cmsClient, err := cms.NewClient(c)
	if err != nil {
		return nil, err
	}
	//cmsClient.SetTransport(rt)
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return &MetricClient{cloudID, cmsClient, logger}, nil
}

//func (c *MetricClient) SetTransport(rate int) {
//	rt := ratelimit.New(rate)
//	c.cms.SetTransport(rt)
//}

// retrive get datapoints for metric
func (c *MetricClient) retrive(sub string, name string, period string) ([]Datapoint, error) {
	req := &cms.DescribeMetricLastRequest{}
	req.Namespace = tea.String(sub)
	req.MetricName = tea.String(name)
	req.Period = tea.String(period)
	data, tryErr := func()(_d []Datapoint, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
				_d = nil
			}
		}()
		resp, err := c.cms.DescribeMetricLast(req)
		if err != nil {
			return nil, err
		}
		var datapoints []Datapoint
		//util.ToJSONString(tea.ToMap(resp))
		_ = json.Unmarshal([]byte(*resp.Body.Datapoints), &datapoints)
		return datapoints, nil
	}()
	if tryErr != nil {
		return nil, tryErr
	}
	return data, nil
}

// Collect do collect metrics into channel
func (c *MetricClient) Collect(namespace string, sub string, m *config.Metric, ch chan<- prometheus.Metric) {
	if m.Name == "" {
		level.Warn(c.logger).Log("msg", "metric name must been set")
		return
	}

	datapoints, err := c.retrive(sub, m.Name, m.Period)
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to retrive datapoints", "cloudID", c.cloudID, "namespace", sub, "name", m.String(), "error", err)
		return
	}
	for _, dp := range datapoints {
		val := dp.Get(m.Measure)
		ch <- prometheus.MustNewConstMetric(
			m.Desc(namespace, sub, dp.Labels()...),
			prometheus.GaugeValue,
			val,
			append(dp.Values(m.Dimensions...), c.cloudID)...,
		)
	}
}

func filterNamespaces(namespaces ...string) []string {
	if len(namespaces) == 0 {
		return allNamesOfNamespaces()
	}
	filters := make([]string, 0)
	for _, ns := range namespaces {
		if ns == "all" {
			return allNamesOfNamespaces()
		}
		if _, ok := allNamespaces[ns]; ok {
			filters = append(filters, ns)
		}
	}
	return filters
}
