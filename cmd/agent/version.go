package agent

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "show DigitalOcean GPU Metrics Agent version info",
		Long:    "show the version of the DigitalOcean GPU Metrics Agent",
		Example: "doGPUMetricsAgent version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf(`DigitalOcean GPU Metrics Agent:
		version                     : %s
		dcgm-exporter version       : %s
		build date                  : %s
		go version                  : %s
		go compiler                 : %s
		platform                    : %s/%s
`, version, dcgmExporterVersion, buildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)

			return nil
		},
	}
)

func init() {
	rootCommand.AddCommand(versionCmd)
}
