package pkg

import (
	"testing"

	"github.com/NVIDIA/dcgm-exporter/pkg/dcgmexporter"
	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestGetCounters(t *testing.T) {
	config := &dcgmexporter.Config{}

	expectedCounters := sets.New[dcgm.Short]()
	for fieldId, _ := range defaultCounters {
		expectedCounters.Insert(fieldId)
	}

	expectedExporterCounters := sets.New[dcgm.Short]()
	for fieldId, _ := range defaultExporterAddedCounters {
		expectedExporterCounters.Insert(fieldId)
	}

	counterSet, err := getCounters(config)
	if err != nil {
		t.Errorf("expected no error, but got: %s", err.Error())
	}

	dcgmCounters := sets.New[dcgm.Short]()
	for _, counter := range counterSet.DCGMCounters {
		dcgmCounters.Insert(counter.FieldID)
	}

	difference := expectedCounters.Difference(dcgmCounters)
	if len(difference) != 0 {
		t.Errorf("difference between expected and actual counters: %v", difference)
	}

	exporterCounters := sets.New[dcgm.Short]()
	for _, counter := range counterSet.ExporterCounters {
		exporterCounters.Insert(counter.FieldID)
	}

	differenceExporterCounters := expectedExporterCounters.Difference(exporterCounters)
	if len(differenceExporterCounters) != 0 {
		t.Errorf("difference between expected and actual counters: %v. Expected: %v, got: %v", differenceExporterCounters, expectedExporterCounters, exporterCounters)
	}
}
