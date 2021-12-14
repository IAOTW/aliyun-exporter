package handler

import (
	"fmt"
	"github.com/IAOTW/aliyun-exporter/pkg/client"
	"github.com/IAOTW/aliyun-exporter/pkg/collector"
	"github.com/IAOTW/aliyun-exporter/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"sigs.k8s.io/yaml"
)

// Handler http metrics handler
type Handler struct {
	logger log.Logger
	server *http.Server
}

// New create http handler
func New(addr string, logger log.Logger, rate int, cfg *config.Config, c map[string]prometheus.Collector) (*Handler, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		logger: logger,
		server: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>Aliyun Exporter</title>
            <style>
            label{
            display:inline-block;
            width:160px;
            }
            form label {
            margin: 10px;
            }
            form input {
            margin: 10px;
            }
            </style>
            </head>
            <body>
            <h1>Aliyun Exporter</h1>
            <form action="/monitors">
            <label>tenantId:</label> <input type="text" name="tenant" placeholder="" value="tenant001" style="width:210px" required><br>
            <label>accessKey:</label> <input type="text" name="accessKey" placeholder="" value="" style="width:210px" required><br>
            <label>accessKeySecret:</label> <input type="text" name="accessKeySecret" placeholder="" value="" style="width:210px" required><br>
            <label>regionId:</label> <input type="text" name="regionId" placeholder="" value="cn-hangzhou" style="width:210px"><br>
            <input type="submit" value="Submit">
            </form>
						<p><a href="/config">Config</a></p>
            </body>
            </html>`))
	})
	http.HandleFunc("/monitors", func(w http.ResponseWriter, r *http.Request) {
		handlerMonitors(w, r, logger, rate, cfg)
	})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		handlerMetrics(w, r, c)
	})
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		c, err := yaml.Marshal(cfg)
		if err != nil {
			level.Error(logger).Log("msg", "Error marshaling configuration", "err", err)
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(c)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Service is UP"))
	})
	return h, nil
}

func handlerMetrics(w http.ResponseWriter, r *http.Request, c map[string]prometheus.Collector) {
	registry := prometheus.NewRegistry()
	for cloudId, _ := range c {
		registry.MustRegister(c[cloudId])
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func handlerMonitors(w http.ResponseWriter, r *http.Request, logger log.Logger, rate int, cfg *config.Config) {
	query := r.URL.Query()

	cloudId := query.Get("tenant")
	if len(query["tenant"]) != 1 || cloudId == "" {
		http.Error(w, "'tenant' parameter must be specified once", 400)
		level.Error(logger).Log("'tenant' parameter must be specified once")
		return
	}

	accessKey := query.Get("accessKey")
	if len(query["accessKey"]) != 1 || accessKey == "" {
		http.Error(w, "'accessKey' parameter must only be specified once", 400)
		level.Error(logger).Log("'accessKey' parameter must be specified once")
		return
	}

	accessKeySecret := query.Get("accessKeySecret")
	if len(query["accessKeySecret"]) != 1 || accessKeySecret == "" {
		http.Error(w, "'accessKeySecret' parameter must only be specified once", 400)
		level.Error(logger).Log("'accessKeySecret' parameter must be specified once")
		return
	}

	regionId := query.Get("regionId")
	if len(query["regionId"]) != 1 || regionId == "" {
		http.Error(w, "'regionId' parameter must only be specified once", 400)
		level.Error(logger).Log("'regionId' parameter must be specified once")
		return
	}
	if cfg.Credentials == nil {
		cfg.Credentials = make(map[string]config.Credential)
	}
	cfg.Credentials[cloudId] = config.Credential{
		accessKey,
		accessKeySecret,
		regionId,
	}
	cfg.SetDefaults()
	cli, err := client.NewMetricClient(cloudId, accessKey, accessKeySecret, regionId, logger)
	if err != nil {
		level.Debug(logger).Log("msg", "Client err", err)
	}
	c := collector.NewCloudMonitorCollectorFromURL(cli, cloudId, cfg, rate, logger)
	registry := prometheus.NewRegistry()
	for i, _ := range c {
		registry.MustRegister(c[i])
	}

	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	return
}

// Run start server
func (h *Handler) Run() error {
	level.Info(h.logger).Log("msg", "Starting metric handler", "port", h.server.Addr)
	fmt.Println("msg", "Starting metric handler", "port", h.server.Addr)
	return h.server.ListenAndServe()
}
