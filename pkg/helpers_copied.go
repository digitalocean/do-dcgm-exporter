package pkg

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/NVIDIA/dcgm-exporter/pkg/dcgmexporter"
	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/*
	This file contains verbatim copied code from the dcgm-exporter, mainly from the file pkg/cmd/app.go
*/

// fillProfilingConfigMetricGroups gets the profiling metric groups supported by the GPU
// Metrics that can be watched concurrently will have different .majorId fields in their dcgmProfMetricGroupInfo_t
// - fields with IDs 1001 - 1033 are general profiling fields
// - fields with IDS 1040 - 1075 are NVLink profiling fields
// - For more information, see: https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/feature-overview.html#multiplexing-of-profiling-counters
// - copied from: https://github.com/NVIDIA/dcgm-exporter/blob/b4552f0bb78fe5b2cd7a17d048ee49abc1c2d926/pkg/cmd/app.go#L481
func fillProfilingConfigMetricGroups(config *dcgmexporter.Config) {
	var groups []dcgm.MetricGroup
	groups, err := dcgm.GetSupportedMetricGroups(0)
	if err != nil {
		config.CollectDCP = false
		logrus.Info("Not collecting DCP metrics: ", err)
	} else {
		logrus.Info("Collecting DCP Metrics")
		config.MetricGroups = groups
	}
}

// enableDCGMExpXIDErrorsCountCollector instantiates the special XID error collector
// - copied from: https://github.com/NVIDIA/dcgm-exporter/blob/b4552f0bb78fe5b2cd7a17d048ee49abc1c2d926/pkg/cmd/app.go#L395
func enableDCGMExpXIDErrorsCountCollector(cs *dcgmexporter.CounterSet, fieldEntityGroupTypeSystemInfo *dcgmexporter.FieldEntityGroupTypeSystemInfo, hostname string, config *dcgmexporter.Config, cRegistry *dcgmexporter.Registry) error {
	if dcgmexporter.IsDCGMExpXIDErrorsCountEnabled(cs.ExporterCounters) {
		item, exists := fieldEntityGroupTypeSystemInfo.Get(dcgm.FE_GPU)
		if !exists {
			return fmt.Errorf("%s collector cannot be initialized", dcgmexporter.DCGMXIDErrorsCount.String())
		}

		xidCollector, err := dcgmexporter.NewXIDCollector(cs.ExporterCounters, hostname, config, item)
		if err != nil {
			return errors.Wrap(err, "failed to instantiate XID collector")
		}

		cRegistry.Register(xidCollector)

		logrus.Debugf("%s collector initialized", dcgmexporter.DCGMXIDErrorsCount.String())
	}
	return nil
}

// enableDCGMExpClockEventsCount instantiates the special clock_events collector
// - copied from: https://github.com/NVIDIA/dcgm-exporter/blob/b4552f0bb78fe5b2cd7a17d048ee49abc1c2d926/pkg/cmd/app.go#L377
func enableDCGMExpClockEventsCount(cs *dcgmexporter.CounterSet, fieldEntityGroupTypeSystemInfo *dcgmexporter.FieldEntityGroupTypeSystemInfo, hostname string, config *dcgmexporter.Config, cRegistry *dcgmexporter.Registry) error {
	if dcgmexporter.IsDCGMExpClockEventsCountEnabled(cs.ExporterCounters) {
		item, exists := fieldEntityGroupTypeSystemInfo.Get(dcgm.FE_GPU)
		if !exists {
			return fmt.Errorf("%s collector cannot be initialized", dcgmexporter.DCGMClockEventsCount.String())
		}
		clocksThrottleReasonsCollector, err := dcgmexporter.NewClockEventsCollector(
			cs.ExporterCounters, hostname, config, item)
		if err != nil {
			return errors.Wrap(err, "failed to instantiate clock events collector")
		}

		cRegistry.Register(clocksThrottleReasonsCollector)

		logrus.Debugf("%s collector initialized", dcgmexporter.DCGMClockEventsCount.String())
	}
	return nil
}

// newOSWatcher sets up a watch for OS signals
// - copied from: https://github.com/NVIDIA/dcgm-exporter/blob/b4552f0bb78fe5b2cd7a17d048ee49abc1c2d926/pkg/cmd/app.go#L269
func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}
