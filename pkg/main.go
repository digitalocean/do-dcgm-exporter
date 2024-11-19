package pkg

import (
	"bytes"
	"fmt"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/NVIDIA/dcgm-exporter/pkg/dcgmexporter"
	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	httpclient "github.com/digitalocean/do-dcgm-exporter/pkg/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/*
	In general in dcgm, to get any field (e.g DCGM_FI_PROF_DRAM_ACTIVE) value you need to:

	1. Initialize dcgm client
	2. Create group
	3. Add devices to the group
	4. Create fields group
	4. Add fields to the group
	5. Enable watches for the field group and device group
	6. Get the latest values for the field group and the device group

	That's all taken care of by the dcgm-exporter code using go-dcgm bindings in the collectors{GPU Collector, NVLink Collector, NVSwitch Collector)}.
*/

const (
	// internalProxyURL is the address of the DO proxy serving an endpoint to receive GPU metrics
	internalProxyURL = "http://169.254.169.254"
	// internalProxyPort is the port of the DO proxy serving an endpoint to receive GPU metrics
	internalProxyPort = 80
	// internalProxyPath is the API path of the DO proxy serving an endpoint to receive GPU metrics
	internalProxyPath = "v1/gpu_metrics"
)

// expMetricsFormat is the go template to render prometheus plaintext format metrics from dcgm_exporter "MetricsByCounter" format
// copied from: https://github.com/NVIDIA/dcgm-exporter/blob/6d499c68a93677e8ccafe857427b09d4617ebfae/pkg/dcgmexporter/expcollector.go#L32
// - reason: not exported field which we need when intercepting metrics "MetricsByCounter" format to convert it to Prometheus plaintext to be pushed to DO proxy
var expMetricsFormat = `

{{- range $counter, $metrics := . -}}
# HELP {{ $counter.FieldName }} {{ $counter.Help }}
# TYPE {{ $counter.FieldName }} {{ $counter.PromType }}
{{- range $metric := $metrics }}
{{ $counter.FieldName }}{gpu="{{ $metric.GPU }}",{{ $metric.UUID }}="{{ $metric.GPUUUID }}",pci_bus_id="{{ $metric.GPUPCIBusID }}",device="{{ $metric.GPUDevice }}",modelName="{{ $metric.GPUModelName }}"{{if $metric.MigProfile}},GPU_I_PROFILE="{{ $metric.MigProfile }}",GPU_I_ID="{{ $metric.GPUInstanceID }}"{{end}}{{if $metric.Hostname }},Hostname="{{ $metric.Hostname }}"{{end}}

{{- range $k, $v := $metric.Labels -}}
	,{{ $k }}="{{ $v }}"
{{- end -}}
{{- range $k, $v := $metric.Attributes -}}
	,{{ $k }}="{{ $v }}"
{{- end -}}

} {{ $metric.Value -}}
{{- end }}
{{ end }}`

// getExpMetricTemplate is the go template to render plaintext prometheus formatted metrics
var getExpMetricTemplate = sync.OnceValue(func() *template.Template {
	return template.Must(template.New("expMetrics").Parse(expMetricsFormat))
})

// NewGPUMetricsAgent creates and returns a new GPUMetricsAgent
func NewGPUMetricsAgent(additionalFieldsPath string, showDebugLogs bool) (*GPUMetricsAgent, error) {
	proxyClient := httpclient.NewHTTP(5 * time.Second)

	dcgmExporterConfig := dcgmexporter.Config{
		// additional fields that can be configured by the user. But can;t overwrite default fields
		CollectorsFile: additionalFieldsPath,
		Address:        ":9401",
		// how often the value of watched fields is read via dcgm (unit in milliseconds)
		CollectInterval: 20000, // every 20s
		Kubernetes:      false,
		CollectDCP:      true, // we want to collect profiling metrics
		UseRemoteHE:     true, // always use pre-installed standalone dcgm to allow customers to run their own dcgm-exporter
		RemoteHEInfo:    "localhost:5555",
		GPUDevices: dcgmexporter.DeviceOptions{
			Flex: true,
		},
		SwitchDevices: dcgmexporter.DeviceOptions{
			Flex: true,
		},
		CPUDevices: dcgmexporter.DeviceOptions{
			Flex: true,
		},
		Debug: showDebugLogs,
		// the time window for the dcgm-exporters clock_events_collector exposing clock throttling reasons via the DCGM_EXP_CLOCK_EVENTS_COUNT metric
		// configured to be equivalent to the collection interval
		ClockEventsCountWindowSize: int((20 * time.Second).Milliseconds()),
		// the time window for the dcgm-exporters xid_collector exposing XID errors via the DCGM_FI_DEV_XID_ERRORS metric
		// configured to be equivalent to the collection interval
		XIDCountWindowSize: int((20 * time.Second).Milliseconds()),
	}

	return &GPUMetricsAgent{
		ProxyClient:        proxyClient,
		DcgmExporterConfig: &dcgmExporterConfig,
	}, nil
}

func (a GPUMetricsAgent) Run() error {
restart:

	cleanup, err := connectToRemoteDCGM(a.DcgmExporterConfig)
	defer func(clean func()) {
		if clean != nil {
			clean()
		}
	}(cleanup)

	if err != nil {
		return errors.Wrap(err, "failed to connect to remote dcgm (nv-hostengine)")
	}

	dcgm.FieldsInit()
	defer dcgm.FieldsTerm()

	logrus.Info("DCGM initialized successfully!")

	fillProfilingConfigMetricGroups(a.DcgmExporterConfig)

	cs, err := getCounters(a.DcgmExporterConfig)
	if err != nil {
		return fmt.Errorf("failed to collect DCGM fields/counters to watch: %s", err.Error())
	}

	fieldEntityGroupTypeSystemInfo := getFieldEntityGroupTypeSystemInfo(cs, a.DcgmExporterConfig)

	hostname, err := dcgmexporter.GetHostname(a.DcgmExporterConfig)
	if err != nil {
		return err
	}

	// the pipeline has reference to all required dcgm-exporter collectors (depends on the counters/watched fields, but typically GPU Collector, NVLink Collector, NVSwitch Collector)
	pipeline, cleanup, err := dcgmexporter.NewMetricsPipeline(a.DcgmExporterConfig,
		cs.DCGMCounters,
		hostname,
		dcgmexporter.NewDCGMCollector,
		fieldEntityGroupTypeSystemInfo,
	)
	defer cleanup()
	if err != nil {
		return errors.Wrap(err, "failed to create metrics pipeline")
	}

	// the registry is a wrapper for the two special collectors {xid_collector, clock_events_collector}.
	// - exposes a Gather() function that call GetMetrics() on both collectors and then aggregates the results
	// - calling Gather() is the mechanism how we obtain the metrics for DCGM_EXP_XID_ERRORS_COUNT and DCGM_EXP_CLOCK_EVENTS_COUNT
	cRegistry := dcgmexporter.NewRegistry()

	// enable XID error collector via the registry
	// - exports prometheus metric: DCGM_EXP_XID_ERRORS_COUNT
	if err := enableDCGMExpXIDErrorsCountCollector(cs, fieldEntityGroupTypeSystemInfo, hostname, a.DcgmExporterConfig, cRegistry); err != nil {
		return err
	}

	// enable collection of clock throttling reasons by resolving bitmask of dcgm field https://docs.nvidia.com/datacenter/dcgm/latest/dcgm-api/dcgm-api-field-ids.html#c.DCGM_FI_DEV_CLOCK_THROTTLE_REASONS
	// - exports prometheus metric: DCGM_EXP_CLOCK_EVENTS_COUNT
	if err := enableDCGMExpClockEventsCount(cs, fieldEntityGroupTypeSystemInfo, hostname, a.DcgmExporterConfig, cRegistry); err != nil {
		return err
	}

	defer func() {
		cRegistry.Cleanup()
	}()

	// channel with 10 plaintext prometheus metrics buffered to be consumed by a reader
	metricsChannel := make(chan string, 10)

	var wg sync.WaitGroup
	stop := make(chan interface{})

	// add to wait-group for pipeline
	wg.Add(1)

	// run the pipeline, invoking all regular collectors {GPU Collector, NVLink Collector, NVSwitch Collector}
	// - each collector returns a slice of metrics for each counter (map[Counter][]Metric).
	// - the pipeline then converts those to the prometheus plain-text format.
	// - the pipeline aggregates the prometheus plain-text metrics of all collectors and sends it out via the metricsChannel
	go pipeline.Run(metricsChannel, stop, &wg)

	// channel with 10 plaintext prometheus metrics buffered to be consumed by the metrics server
	// - fed from the actual metrics channel
	metricsServerChannel := make(chan string, 10)

	// Continuously read from the pipeline + registry to forward metrics to the DO proxy
	//  - IDEA: instead of passing though the metricsChannel directly to the MetricServer, we already read from it here
	//  - Reason: to push metrics to the rmetdataproxy, we need to have access to the metrics gathered by both the pipeline and the registry
	//  - Next: Forward the exact same metrics to the metric server to expose on /metrics for customers to query (just like the dcgm-exporter does)
	go func() {
		for {
			select {
			case <-stop:
				return
			case plaintextMetrics := <-metricsChannel:
				metricsBuffer := bytes.Buffer{}
				metricsBuffer.Write([]byte(plaintextMetrics))

				// forward metrics to the dcgm-exporter's metrics server
				metricsServerChannel <- plaintextMetrics

				// the pipeline sends on the metrics channel every config.CollectInterval(20s) seconds - that's the same timeframe as the XID + clock_events collector window
				// Hence, we can from the registry and get accurate metrics over the last time window
				metrics, err := cRegistry.Gather()
				if err != nil {
					logrus.Error("failed to gather metrics from the registry(XID Collector, clock_events collector)")
					continue
				}

				var tmpBuffer bytes.Buffer

				tmpl := getExpMetricTemplate()
				if err := tmpl.Execute(&tmpBuffer, metrics); err != nil {
					logrus.Error("failed to template metrics from the registry(XID Collector, clock_events collector) into prometheus plaintext format")
					continue
				}

				// append metrics to buffer
				metricsBuffer.Write(tmpBuffer.Bytes())

				// finally send the metrics to internal DO systems
				go func(buf bytes.Buffer) {
					if err := a.forwardMetricsToProxy(&buf); err != nil {
						logrus.Error(err.Error())
						return
					}

					logrus.Debug("Successfully forwarded metrics")
				}(metricsBuffer)
			}
		}
	}()

	// serve a /metrics endpoint just like the dcgm-exporter does
	// - stores duplicate metrics data in plaintext, but that's fine
	// - every time the endpoint is hit, reads from the pre-gather metrics consumed from metricsServerChannel, but also invokes the registry (hence there is additional overhead)
	server, cleanup, err := dcgmexporter.NewMetricsServer(a.DcgmExporterConfig, metricsServerChannel, cRegistry)
	defer cleanup()
	if err != nil {
		return err
	}

	// add to wait-group for metrics server
	wg.Add(1)
	go server.Run(stop, &wg)

	sigs := newOSWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	// wait before terminating: wait for one of the OS signals to be delivered to the process
	sig := <-sigs

	// signal termination to {pipeline, metrics server}
	close(stop)

	// wait for {pipeline, metrics server} to have terminated, or 2 seconds, whatever comes earlier
	err = dcgmexporter.WaitWithTimeout(&wg, time.Second*2)
	if err != nil {
		return err
	}

	if sig == syscall.SIGHUP {
		goto restart
	}

	return nil
}
