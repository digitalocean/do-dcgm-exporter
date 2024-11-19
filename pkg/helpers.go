package pkg

import (
	"github.com/NVIDIA/dcgm-exporter/pkg/dcgmexporter"
	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"github.com/sirupsen/logrus"
)

// connectToRemoteDCGM connects to the standalone already running nv-hostengine process (this is a prerequisite)
func connectToRemoteDCGM(config *dcgmexporter.Config) (func(), error) {
	logrus.Info("Attempting to connect to remote hostengine at ", config.RemoteHEInfo)
	cleanup, err := dcgm.Init(dcgm.Standalone, config.RemoteHEInfo, "0")
	if err != nil {
		return cleanup, err
	}
	return cleanup, nil
}

// getCounters returns a set of counters for which we collect metrics.
// These could be either DCGM fields or dcgm-exporter added metrics
// - DCGM Counters: fields from https://docs.nvidia.com/datacenter/dcgm/latest/dcgm-api/dcgm-api-field-ids.html
// - Exporter Counters: counters added by dcgm-exporter that don't exist in dcgm {DCGM_EXP_CLOCK_EVENTS_COUNT, DCGM_EXP_XID_ERRORS_COUNT, label-type exporter metrics {DCGM_FI_DRIVER_VERSION, DCGM_FI_NVML_VERSION, ...}}
func getCounters(config *dcgmexporter.Config) (*dcgmexporter.CounterSet, error) {
	counterSetFromFile := &dcgmexporter.CounterSet{}
	var err error

	// read counters from configured CSV file
	if len(config.CollectorsFile) > 0 {
		counterSetFromFile, err = dcgmexporter.GetCounterSet(config)
		if err != nil {
			return nil, err
		}

		// Copy labels from DCGM Counters to ExporterCounters
		for i := range counterSetFromFile.DCGMCounters {
			if counterSetFromFile.DCGMCounters[i].PromType == "label" {
				counterSetFromFile.ExporterCounters = append(counterSetFromFile.ExporterCounters, counterSetFromFile.DCGMCounters[i])
			}
		}
	}

	counterSet := dcgmexporter.CounterSet{}

	// add default counters for dcgm fields
	for _, defaultCounter := range defaultCounters {
		counterSet.DCGMCounters = append(counterSet.DCGMCounters, defaultCounter)
	}

	// add default counters for dcgm_exporter only fields (don't exist in dcgm)
	for _, defaultCounter := range defaultExporterAddedCounters {
		counterSet.ExporterCounters = append(counterSet.ExporterCounters, defaultCounter)
	}

	// add additional counters from the configuration file
	for _, counter := range counterSetFromFile.DCGMCounters {
		_, ok := defaultCounters[counter.FieldID]
		// avoid adding duplicates
		if !ok {
			counterSet.DCGMCounters = append(counterSet.DCGMCounters, counter)
		}
	}

	for _, counter := range counterSetFromFile.ExporterCounters {
		_, ok := defaultCounters[counter.FieldID]
		// avoid adding duplicates
		if !ok {
			counterSet.ExporterCounters = append(counterSet.ExporterCounters, counter)
		}
	}

	return &counterSet, nil
}

// getFieldEntityGroupTypeSystemInfo creates a mapping {dcgm.FIELD_ENTITY_GROUP{FE_GPU,FE_SWITCH,FE_LINK} -> (system_info such as GPUs on the system, field_ids to watch for that group extracted from counters)}
// - under the hood, discovers the available hardware  (GPUs, NVLinks, NVSwitches, CPUs) to obtain the system info.
// - adapted from: https://github.com/NVIDIA/dcgm-exporter/blob/b4552f0bb78fe5b2cd7a17d048ee49abc1c2d926/pkg/cmd/app.go#L413
func getFieldEntityGroupTypeSystemInfo(cs *dcgmexporter.CounterSet, config *dcgmexporter.Config) *dcgmexporter.FieldEntityGroupTypeSystemInfo {
	fieldEntityGroupTypeSystemInfo := dcgmexporter.NewEntityGroupTypeSystemInfo(cs.DCGMCounters, config)

	for _, egt := range dcgmexporter.FieldEntityGroupTypeToMonitor {
		// Discovers physical devices (GPUs, NVLinks, NVSwitches, CPUs) using dcgm_exporter.system_info -> dcgm
		// - uses system_info from dcgm_exporter that cannot be mocked due to unexported fields
		err := fieldEntityGroupTypeSystemInfo.Load(egt)
		if err != nil {
			// Typically errors indicate that we are not watching any fields in the CPU groups
			// - INFO[0001] Not collecting CPU metrics; no fields to watch for device type: 7
			// - INFO[0001] Not collecting CPU Core metrics; no fields to watch for device type: 8
			logrus.Infof("Not collecting %s metrics: %s", egt.String(), err)
		}
	}
	return fieldEntityGroupTypeSystemInfo
}
