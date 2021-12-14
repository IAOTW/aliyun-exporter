package cmd

import (
	"fmt"
	"github.com/IAOTW/aliyun-exporter/pkg/client"
	"github.com/IAOTW/aliyun-exporter/pkg/collector"
	"github.com/IAOTW/aliyun-exporter/pkg/config"
	"github.com/IAOTW/aliyun-exporter/pkg/handler"
	"github.com/IAOTW/aliyun-exporter/version"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

const AppName = "cloudmonitor"

// NewRootCommand create root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           AppName,
		Short:         "Exporter for aliyun cloudmonitor",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(newServeMetricsCommand())
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newListMetricNamespacesCommand())
	return cmd
}

func newServeMetricsCommand() *cobra.Command {
	o := &options{
		so: &serveOption{},
	}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve HTTP metrics handler",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return o.Complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Parse(o.so.configFile)
			if err != nil {
				return err
			}
			cms, err := collector.NewCloudMonitorCollector(AppName, cfg, o.rateLimit, logger)
			if err != nil {
				return err
			}
			h, err := handler.New(o.so.listenAddress, logger, o.rateLimit, cfg, cms)
			if err != nil {
				return err
			}
			return h.Run()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.Version())
		},
	}
}

func newListMetricNamespacesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-metrics",
		Short: "List avaliable namespaces of metrics",
		Run: func(_ *cobra.Command, _ []string) {
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
			fmt.Fprintln(w, "NAMESPACE\tDESCRIPTION")
			for name, desc := range client.AllNamespaces() {
				fmt.Fprintf(w, "%s\t%s\n", name, desc)
			}
			w.Flush()
		},
	}
}
