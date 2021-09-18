package handler

import (
	"fmt"
	"net"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler http metrics handler
type Handler struct {
	logger   log.Logger
	server   *http.Server
	registry *prometheus.Registry
}

// New create http handler
func New(addr string, metricPath string, logger log.Logger, c ...map[string]prometheus.Collector) (*Handler, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		logger: logger,
		server: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
		registry: prometheus.NewRegistry(),
	}

	// metric collectors
	h.MustRegister(c...)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(fmt.Sprintf(`
		<h1>Aliyun CloudMonitor Exporter, visit <a href="%s">HERE</a> for metrics</h1>`, metricPath)))
	})
	mux.Handle(metricPath, promhttp.InstrumentMetricHandler(h.registry, promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{})))
	h.server.Handler = mux
	return h, nil
}

// MustRegister register collector or die
func (h *Handler) MustRegister(c ...map[string]prometheus.Collector) {
	for _, collectors := range c {
		for _, collector := range collectors {
			h.registry.MustRegister(collector)
		}
	}
}

// Run start server
func (h *Handler) Run() error {
	level.Info(h.logger).Log("msg", "Starting metric handler", "port", h.server.Addr)
	fmt.Println("msg", "Starting metric handler", "port", h.server.Addr)
	return h.server.ListenAndServe()
}
