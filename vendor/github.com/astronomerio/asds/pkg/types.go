package pkg

import "github.com/prometheus/client_golang/prometheus"

type Resources struct {
	Disk  float64 `json:"disk,omitempty"`
	Mem   float64 `json:"mem,omitempty"`
	GPUs  float64 `json:"gpus,omitempty"`
	CPUs  float64 `json:"cpus,omitempty"`
	Ports string  `json:"ports,omitempty"`
}

type Framework struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name,omitempty"`
	UsedResources    Resources `json:"used_resources,omitempty"`
	OfferedResources Resources `json:"offered_resources,omitempty"`
	Capabilities     []string  `json:"capabilities,omitempty"`
	Hostname         string    `json:"hostname,omitempty"`
	WebUIURL         string    `json:"webui_url,omitempty"`
	Active           bool      `json:"active,omitempty"`
	Connected        bool      `json:"connected,omitempty"`
	Recovered        bool      `json:"recovered,omitempty"`
	TaskStaging      int       `json:"TASK_STAGING,omitempty"`
	TaskStarting     int       `json:"TASK_STARTING,omitempty"`
	TaskRunning      int       `json:"TASK_RUNNING,omitempty"`
	TaskKilling      int       `json:"TASK_KILLING,omitempty"`
	TaskFinished     int       `json:"TASK_FINISHED,omitempty"`
	TaskKilled       int       `json:"TASK_KILLED,omitempty"`
	TaskFailed       int       `json:"TASK_FAILED,omitempty"`
	TaskLost         int       `json:"TASK_LOST,omitempty"`
	TaskError        int       `json:"TASK_ERROR,omitempty"`
	TaskUnreachable  int       `json:"TASK_UNREACHABLE,omitempty"`
	SlaveIds         []string  `json:"slave_ids,omitempty"`
}

type Slaves struct {
	ID                  string            `json:"id,omitempty"`
	Hostname            string            `json:"hostname,omitempty"`
	Port                int               `json:"port,omitempty"`
	Attributes          map[string]string `json:"attributes,omitempty"`
	Pid                 string            `json:"pid,omitempty"`
	RegisteredTime      float64           `json:"registered_time,omitempty"`
	Resources           Resources         `json:"resources,omitempty"`
	UsedResources       Resources         `json:"used_resources,omitempty"`
	OfferedResources    Resources         `json:"offered_resources,omitempty"`
	ReservedResources   Resources         `json:"reserved_resources,omitempty"`
	UnreservedResources Resources         `json:"unreserved_resources,omitempty"`
	Active              bool              `json:"active,omitempty"`
	Version             string            `json:"version,omitempty"`
	TaskStaging         int               `json:"TASK_STAGING,omitempty"`
	TaskStarting        int               `json:"TASK_STARTING,omitempty"`
	TaskRunning         int               `json:"TASK_RUNNING,omitempty"`
	TaskKilling         int               `json:"TASK_KILLING,omitempty"`
	TaskFinished        int               `json:"TASK_FINISHED,omitempty"`
	TaskKilled          int               `json:"TASK_KILLED,omitempty"`
	TaskFailed          int               `json:"TASK_FAILED,omitempty"`
	TaskLost            int               `json:"TASK_LOST,omitempty"`
	TaskError           int               `json:"TASK_ERROR,omitempty"`
	TaskUnreachable     int               `json:"TASK_UNREACHABLE,omitempty"`
	FrameworkIds        []string          `json:"framework_ids,omitempty"`
}
type StateSummary struct {
	Hostname   string      `json:"hostname,omitempty"`
	Cluster    string      `json:"cluster,omitempty"`
	Slaves     []Slaves    `json:"slaves,omitempty"`
	Frameworks []Framework `json:"frameworks,omitempty"`
}

type PrometheusInstrumentation struct {
	ProResourceAvailableCPU  prometheus.GaugeVec
	ProResourceAvailableMem  prometheus.GaugeVec
	ProResourceAvailableDisk prometheus.GaugeVec
	ProResourceTotalCPU      prometheus.Gauge
	ProResourceTotalMem      prometheus.Gauge
	ProResourceTotalDisk     prometheus.Gauge
	ProDeploymentCount       prometheus.CounterVec
	ProProvisionCount        prometheus.Counter
	ProDeprovisionCount      prometheus.Counter
}
