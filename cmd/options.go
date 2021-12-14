package cmd

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var logger log.Logger

// options command line options
type options struct {
	logFmt    string
	logLevel  string
	logFile   string
	rateLimit int
	so        *serveOption
}

func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.logFmt, "log.format", "logfmt", "Output format of log messages. One of: [logfmt, json]")
	cmd.Flags().StringVar(&o.logLevel, "log.level", "info", "Log level")
	cmd.Flags().StringVar(&o.logFile, "log.file", "aliyun-exporter.log", "Log message to file")
	cmd.Flags().IntVar(&o.rateLimit, "rate-limit", 20, "RPS/request per second")
	if o.so != nil {
		o.so.AddFlags(cmd)
	}
}

// Complete do some initialization
func (o *options) Complete() error {
	writer := os.Stdout
	if o.logFile != "" {
		out, err := os.OpenFile(o.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writer = out
	}
	switch o.logFmt {
	case "logfmt":
		logger = log.NewLogfmtLogger(writer)
	case "json":
		logger = log.NewJSONLogger(writer)
	default:
		logger = log.NewNopLogger()
	}
	current := func() time.Time {
		return time.Now()
	}
	logger = log.With(logger, "ts", log.Timestamp(current), "caller", log.DefaultCaller)
	var lvlOp level.Option
	switch strings.ToLower(o.logLevel) {
	case "debug":
		lvlOp = level.AllowDebug()
	case "info":
		lvlOp = level.AllowInfo()
	case "warn", "warning":
		lvlOp = level.AllowWarn()
	case "error":
		lvlOp = level.AllowError()
	default:
		level.Info(logger).Log("msg", "unknown log level, fallback to info", "level", o.logLevel)
		lvlOp = level.AllowInfo()
	}
	logger = level.NewFilter(logger, lvlOp)
	if o.so != nil {
		if err := o.so.Complete(); err != nil {
			return err
		}
	}
	return nil
}

type serveOption struct {
	configFile    string
	listenAddress string
}

func (o *serveOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.configFile, "config", "c", "config.yaml", "Path of config file")
	cmd.Flags().StringVar(&o.listenAddress, "web.listen-address", ":9527", "Address on which to expose metrics and web interface.")
}

func (o *serveOption) Complete() error {
	return nil
}
