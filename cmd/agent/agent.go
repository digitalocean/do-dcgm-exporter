package agent

import (
	"github.com/digitalocean/do-dcgm-exporter/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// version command
	version string
	// dcgmExporterVersion is the dcgm-exporter version the DigitalOcean dcgm-exporter is build on
	dcgmExporterVersion string
	// buildDate is the date when the binary was build
	buildDate string
	// showDebugLogs is a flag to enable debug logs
	showDebugLogs bool
	// additionalFieldsPath is the path to a file containing additional DCGM fields to monitor
	additionalFieldsPath string

	rootCommand = &cobra.Command{
		Use:     "do-dcgm-exporter",
		Short:   "The DigitalOcean DCGM exporter",
		Long:    `The DigitalOcean DCGM exporter collects GPU Metrics via dcgm and pushes them to internal DO systems`,
		Version: version,
		Args: func(cmd *cobra.Command, args []string) error {
			return cmd.ParseFlags(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showDebugLogs {
				logrus.SetLevel(logrus.DebugLevel)
			}

			agent, err := pkg.NewGPUMetricsAgent(additionalFieldsPath, showDebugLogs)
			if err != nil {
				return err
			}

			err = agent.Run()
			if err != nil {
				err = errors.Wrap(err, "failed to run the do-dcgm-exporter")
			}

			return err
		},
		SilenceUsage: true,
	}
)

func init() {
	rootCommand.Flags().BoolVar(
		&showDebugLogs,
		"debug",
		false,
		"show debug logs")

	rootCommand.Flags().StringVar(
		&additionalFieldsPath,
		"collectors", // compatibility with dcgm-exporter
		"",
		"Path to the file, that contains additional DCGM fields to collect. These are fields beyond the default fields that are always collected")

}

func NewCommandStartAgent() *cobra.Command {
	return rootCommand
}
