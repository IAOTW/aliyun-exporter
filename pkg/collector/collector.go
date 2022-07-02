package collector

import (
	"github.com/IAOTW/aliyun-exporter/pkg/client"
	"github.com/IAOTW/aliyun-exporter/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

const AppName = "cloudmonitor"

// cloudMonitor ..
type cloudMonitor struct {
	namespace string
	cfg       *config.Config
	logger    log.Logger
	// sdk client
	client *client.MetricClient
	rate   int
	lock   sync.Mutex
}

// NewCloudMonitorCollector create a new collector for cloud monitor
func NewCloudMonitorCollector(appName string, cfg *config.Config, rate int, logger log.Logger) (map[string]prometheus.Collector, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	cloudMonitors := make(map[string]prometheus.Collector)
	for cloudID, credential := range cfg.Credentials {
		cli, err := client.NewMetricClient(cloudID, credential.AccessKey, credential.AccessKeySecret, credential.Region, logger)
		if err != nil {
			continue
		}
		cloudMonitors[cloudID] = &cloudMonitor{
			namespace: appName,
			cfg:       cfg,
			logger:    logger,
			client:    cli,
			rate:      rate,
		}
	}
	return cloudMonitors, nil
}

// NewCloudMonitorCollectorFromURL create a new collector from HTTP Request URL for cloud monitor
func NewCloudMonitorCollectorFromURL(cli *client.MetricClient, cloudID string, cfg *config.Config, rate int, logger log.Logger) map[string]prometheus.Collector {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	collectors := make(map[string]prometheus.Collector)
	collectors[cloudID] = &cloudMonitor{
		namespace: AppName,
		cfg:       cfg,
		logger:    logger,
		client:    cli,
		rate:      rate,
	}
	return collectors
}

func (m *cloudMonitor) Describe(ch chan<- *prometheus.Desc) {
}

func (m *cloudMonitor) Collect(ch chan<- prometheus.Metric) {
	m.lock.Lock()
	defer m.lock.Unlock()

	wg := &sync.WaitGroup{}
	// do collect
	//m.client.SetTransport(m.rate)
	for sub, metrics := range m.cfg.Metrics {
		for i := range metrics {
			wg.Add(1)
			go func(namespace string, metric *config.Metric) {
				defer wg.Done()
				m.client.Collect(m.namespace, namespace, metric, ch)
			}(sub, metrics[i])
		}
	}
	wg.Wait()
}
