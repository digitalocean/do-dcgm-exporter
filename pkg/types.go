package pkg

import (
	"github.com/NVIDIA/dcgm-exporter/pkg/dcgmexporter"
	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	httpclient "github.com/digitalocean/do-dcgm-exporter/pkg/client"
)

// GPUMetricsAgent obtains prometheus-style metrics from a scrapable dcgm-exporter, filters for a whitelisted set of metrics,
// and then pushes the metrics to an internal DO system
type GPUMetricsAgent struct {
	// ProxyClient sends HTTP POST requests to internal DO systems
	ProxyClient httpclient.HTTPClient

	// DcgmExporterConfig is the configuration of the underlying DCGM exporter
	DcgmExporterConfig *dcgmexporter.Config
}

var (
	// defaultCounters are a default set of counters that are always monitored
	// - for all possible DCGM fields that can be configured in the dcgm-exporter, please see: https://docs.nvidia.com/datacenter/dcgm/latest/dcgm-api/dcgm-api-field-ids.html.
	defaultCounters = map[dcgm.Short]dcgmexporter.Counter{
		1: {
			FieldID:   1,
			FieldName: "DCGM_FI_DRIVER_VERSION",
			PromType:  "label", // becomes a label on every metric
			Help:      "The NVIDIA driver version.",
		},
		100: {
			FieldID:   100,
			FieldName: "DCGM_FI_DEV_SM_CLOCK",
			PromType:  "gauge",
			Help:      "SM clock frequency (in MHz).",
		},
		101: {
			FieldID:   101,
			FieldName: "DCGM_FI_DEV_MEM_CLOCK",
			PromType:  "gauge",
			Help:      "Memory clock frequency (in MHz).",
		},
		140: {
			FieldID:   140,
			FieldName: "DCGM_FI_DEV_MEMORY_TEMP",
			PromType:  "gauge",
			Help:      "Memory temperature (in C).",
		},
		150: {
			FieldID:   150,
			FieldName: "DCGM_FI_DEV_GPU_TEMP",
			PromType:  "gauge",
			Help:      "GPU temperature (in C).",
		},
		155: {
			FieldID:   155,
			FieldName: "DCGM_FI_DEV_POWER_USAGE",
			PromType:  "gauge",
			Help:      "Power draw (in W).",
		},
		158: {
			FieldID:   158,
			FieldName: "DCGM_FI_DEV_SLOWDOWN_TEMP",
			PromType:  "gauge",
			Help:      "GPU Slowdown temperature.",
		},
		159: {
			FieldID:   159,
			FieldName: "DCGM_FI_DEV_SHUTDOWN_TEMP",
			PromType:  "gauge",
			Help:      "GPU Shutdown temperature.",
		},
		190: {
			FieldID:   190,
			FieldName: "DCGM_FI_DEV_FAN_SPEED",
			PromType:  "gauge",
			Help:      "Fan speed for the device in percent 0-100",
		},
		191: {
			FieldID:   191,
			FieldName: "DCGM_FI_DEV_PSTATE",
			PromType:  "gauge",
			Help:      "GPU Performance state (P-State) 0-15. 0=highest.",
		},

		// PCIe
		202: {
			FieldID:   202,
			FieldName: "DCGM_FI_DEV_PCIE_REPLAY_COUNTER",
			PromType:  "counter",
			Help:      "Total number of PCIe retries.",
		},

		// Encoder/Decoder utilization
		206: {
			FieldID:   206,
			FieldName: "DCGM_FI_DEV_ENC_UTIL",
			PromType:  "gauge",
			Help:      "Encoder utilization (in %).",
		},
		207: {
			FieldID:   207,
			FieldName: "DCGM_FI_DEV_DEC_UTIL",
			PromType:  "gauge",
			Help:      "Decoder utilization (in %).",
		},

		// Errors and violations
		240: {
			FieldID:   240,
			FieldName: "DCGM_FI_DEV_THERMAL_VIOLATION",
			PromType:  "counter",
			Help:      "Throttling duration due to thermal constraints (in us).",
		},
		241: {
			FieldID:   241,
			FieldName: "DCGM_FI_DEV_POWER_VIOLATION",
			PromType:  "counter",
			Help:      "Throttling duration due to power constraints (in us).",
		},
		242: {
			FieldID:   242,
			FieldName: "DCGM_FI_DEV_SYNC_BOOST_VIOLATION",
			PromType:  "counter",
			Help:      "Throttling duration due to sync-boost constraints (in us).",
		},
		243: {
			FieldID:   243,
			FieldName: "DCGM_FI_DEV_BOARD_LIMIT_VIOLATION",
			PromType:  "counter",
			Help:      "Throttling duration due to board limit constraints (in us).",
		},
		244: {
			FieldID:   244,
			FieldName: "DCGM_FI_DEV_LOW_UTIL_VIOLATION",
			PromType:  "counter",
			Help:      "Throttling duration due to low utilization (in us).",
		},
		245: {
			FieldID:   245,
			FieldName: "DCGM_FI_DEV_RELIABILITY_VIOLATION",
			PromType:  "counter",
			Help:      "Throttling duration due to reliability constraints (in us).",
		},

		// ECC errors
		310: {
			FieldID:   310,
			FieldName: "DCGM_FI_DEV_ECC_SBE_VOL_TOTAL",
			PromType:  "counter",
			Help:      "Total number of single-bit volatile ECC errors.",
		},
		311: {
			FieldID:   311,
			FieldName: "DCGM_FI_DEV_ECC_DBE_VOL_TOTAL",
			PromType:  "counter",
			Help:      "Total number of double-bit volatile ECC errors.",
		},
		312: {
			FieldID:   312,
			FieldName: "DCGM_FI_DEV_ECC_SBE_AGG_TOTAL",
			PromType:  "counter",
			Help:      "Total number of single-bit persistent ECC errors.",
		},
		313: {
			FieldID:   313,
			FieldName: "DCGM_FI_DEV_ECC_DBE_AGG_TOTAL",
			PromType:  "counter",
			Help:      "Total number of double-bit persistent ECC errors.",
		},
		390: {
			FieldID:   390,
			FieldName: "DCGM_FI_DEV_RETIRED_SBE",
			PromType:  "counter",
			Help:      "Number of retired pages because of single bit errors Note: monotonically increasing",
		},
		391: {
			FieldID:   391,
			FieldName: "DCGM_FI_DEV_RETIRED_DBE",
			PromType:  "counter",
			Help:      "Number of retired pages because of double bit errors Note: monotonically increasing",
		},

		// NVLINK
		409: {
			FieldID:   409,
			FieldName: "DCGM_FI_DEV_NVLINK_CRC_FLIT_ERROR_COUNT_TOTAL",
			PromType:  "counter",
			Help:      "Total number of NVLink flow-control CRC errors.",
		},
		419: {
			FieldID:   419,
			FieldName: "DCGM_FI_DEV_NVLINK_CRC_DATA_ERROR_COUNT_TOTAL",
			PromType:  "counter",
			Help:      "Total number of NVLink data CRC errors.",
		},
		429: {
			FieldID:   429,
			FieldName: "DCGM_FI_DEV_NVLINK_REPLAY_ERROR_COUNT_TOTAL",
			PromType:  "counter",
			Help:      "Total number of NVLink retries.",
		},
		439: {
			FieldID:   439,
			FieldName: "DCGM_FI_DEV_NVLINK_RECOVERY_ERROR_COUNT_TOTAL",
			PromType:  "counter",
			Help:      "Total number of NVLink recovery errors.",
		},
		449: {
			FieldID:   449,
			FieldName: "DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL",
			PromType:  "counter",
			Help:      "Total number of NVLink bandwidth counters for all lanes",
		},
		870: {
			FieldID:   870,
			FieldName: "DCGM_FI_DEV_NVSWITCH_LINK_STATUS",
			PromType:  "gauge",
			Help:      "NvLink status {UNKNOWN:-1 OFF:0 SAFE:1 ACTIVE:2 ERROR:3 INACTIVE: 4}",
		},

		// NVSwitch
		858: {
			FieldID:   858,
			FieldName: "DCGM_FI_DEV_NVSWITCH_TEMPERATURE_CURRENT",
			PromType:  "gauge",
			Help:      "NVSwitch current temperature.",
		},
		859: {
			FieldID:   859,
			FieldName: "DCGM_FI_DEV_NVSWITCH_TEMPERATURE_LIMIT_SLOWDOWN",
			PromType:  "gauge",
			Help:      "NVSwitch limit slowdown temperature.",
		},
		860: {
			FieldID:   860,
			FieldName: "DCGM_FI_DEV_NVSWITCH_TEMPERATURE_LIMIT_SHUTDOWN",
			PromType:  "gauge",
			Help:      "NVSwitch limit shutdown temperature.",
		},
		861: {
			FieldID:   861,
			FieldName: "DCGM_FI_DEV_NVSWITCH_THROUGHPUT_TX",
			PromType:  "gauge",
			Help:      "NVSwitch throughput Tx.",
		},
		862: {
			FieldID:   862,
			FieldName: "DCGM_FI_DEV_NVSWITCH_THROUGHPUT_RX",
			PromType:  "gauge",
			Help:      "NVSwitch throughput Rx.",
		},
		856: {
			FieldID:   856,
			FieldName: "DCGM_FI_DEV_NVSWITCH_FATAL_ERRORS",
			PromType:  "counter",
			Help:      "NVSwitch fatal error information. Note: value field indicates the specific SXid reported",
		},
		857: {
			FieldID:   857,
			FieldName: "DCGM_FI_DEV_NVSWITCH_NON_FATAL_ERRORS",
			PromType:  "counter",
			Help:      "NVSwitch non-fatal error information. Note: value field indicates the specific SXid reported",
		},

		// Remapped rows
		393: {
			FieldID:   393,
			FieldName: "DCGM_FI_DEV_UNCORRECTABLE_REMAPPED_ROWS",
			PromType:  "counter",
			Help:      "Number of remapped rows for uncorrectable errors",
		},
		394: {
			FieldID:   394,
			FieldName: "DCGM_FI_DEV_CORRECTABLE_REMAPPED_ROWS",
			PromType:  "counter",
			Help:      "Number of remapped rows for correctable errors",
		},
		395: {
			FieldID:   395,
			FieldName: "DCGM_FI_DEV_ROW_REMAP_FAILURE",
			PromType:  "gauge",
			Help:      "Whether remapping of rows has failed",
		},

		// Memory usage
		254: {
			FieldID:   254,
			FieldName: "DCGM_FI_DEV_FB_USED_PERCENT",
			PromType:  "gauge",
			Help:      "Percentage used of Frame Buffer: ‘Used/(Total - Reserved)’. Range 0.0-1.0",
		},

		// Profiling counters
		1001: {
			FieldID:   1001,
			FieldName: "DCGM_FI_PROF_GR_ENGINE_ACTIVE",
			PromType:  "gauge",
			Help:      "Ratio of time the graphics engine is active.",
		},
		1002: {
			FieldID:   1002,
			FieldName: "DCGM_FI_PROF_SM_ACTIVE",
			PromType:  "gauge",
			Help:      "The ratio of cycles an SM has at least 1 warp assigned.",
		},
		1003: {
			FieldID:   1003,
			FieldName: "DCGM_FI_PROF_SM_OCCUPANCY",
			PromType:  "gauge",
			Help:      "The ratio of number of warps resident on an SM.",
		},
		1004: {
			FieldID:   1004,
			FieldName: "DCGM_FI_PROF_PIPE_TENSOR_ACTIVE",
			PromType:  "gauge",
			Help:      "Ratio of cycles the tensor (HMMA) pipe is active.",
		},
		1005: {
			FieldID:   1005,
			FieldName: "DCGM_FI_PROF_DRAM_ACTIVE",
			PromType:  "gauge",
			Help:      "The ratio of cycles the device memory interface is active sending or receiving data.",
		},
		1006: {
			FieldID:   1006,
			FieldName: "DCGM_FI_PROF_PIPE_FP64_ACTIVE",
			PromType:  "gauge",
			Help:      "Ratio of cycles the fp64 pipes are active.",
		},
		1007: {
			FieldID:   1007,
			FieldName: "DCGM_FI_PROF_PIPE_FP32_ACTIVE",
			PromType:  "gauge",
			Help:      "Ratio of cycles the fp32 pipes are active.",
		},
		1008: {
			FieldID:   1008,
			FieldName: "DCGM_FI_PROF_PIPE_FP16_ACTIVE",
			PromType:  "gauge",
			Help:      "Ratio of cycles the fp16 pipes are active.",
		},
		1009: {
			FieldID:   1009,
			FieldName: "DCGM_FI_PROF_PCIE_TX_BYTES",
			PromType:  "gauge",
			Help:      "The rate of data transmitted over the PCIe bus - including both protocol headers and data payloads - in bytes per second.",
		},
		1010: {
			FieldID:   1010,
			FieldName: "DCGM_FI_PROF_PCIE_RX_BYTES",
			PromType:  "gauge",
			Help:      "The rate of data received over the PCIe bus - including both protocol headers and data payloads - in bytes per second.",
		},
		1011: {
			FieldID:   1011,
			FieldName: "DCGM_FI_PROF_NVLINK_TX_BYTES",
			PromType:  "gauge",
			Help:      "The rate of data transmitted over NVLink not including protocol headers in bytes per second",
		},
		1012: {
			FieldID:   1012,
			FieldName: "DCGM_FI_PROF_NVLINK_RX_BYTES",
			PromType:  "gauge",
			Help:      "The rate of data received over NVLink not including protocol headers in bytes per second",
		},

		// other configuration
		66: {
			FieldID:   66,
			FieldName: "DCGM_FI_DEV_PERSISTENCE_MODE",
			PromType:  "gauge",
			Help:      "Persistence mode for the device Boolean: 0 is disabled 1 is enabled",
		},

		// required by xid_collector to compute DCGM_EXP_XID_ERRORS_COUNT
		230: {
			FieldID:   230,
			FieldName: "DCGM_FI_DEV_XID_ERRORS",
			PromType:  "gauge",
			Help:      "Value of the last XID error encountered.",
		},
		// required by clock_event_exporter to compute DCGM_EXP_CLOCK_EVENTS_COUNT
		112: {
			FieldID:   112,
			FieldName: "DCGM_FI_DEV_CLOCK_THROTTLE_REASONS",
			PromType:  "gauge",
			Help:      "Current clock throttle reasons (bitmask of DCGM_CLOCKS_THROTTLE_REASON_*)",
		},
	}

	// these counters are added by the dcgm_exporter and don't exist as dcgm fields
	defaultExporterAddedCounters = map[dcgm.Short]dcgmexporter.Counter{
		9001: {
			FieldID:   9001,
			FieldName: "DCGM_EXP_XID_ERRORS_COUNT",
			PromType:  "gauge",
			Help:      "Count of XID errors within a 20s time window",
		},
		9002: {
			FieldID:   9002,
			FieldName: "DCGM_EXP_CLOCK_EVENTS_COUNT",
			PromType:  "gauge",
			Help:      "Count of clock events within a 20s time window.",
		},
	}
)
