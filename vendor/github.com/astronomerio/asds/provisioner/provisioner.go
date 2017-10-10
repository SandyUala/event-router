package provisioner

type Provisioner interface {
	Provision(provRequest *ProvisionRequest) (*ProvisionResponse, error)
	ListProvisions() (*ProvisionResponse, error)
	GetProvision(provRequest *ProvisionRequest) (*ProvisionResponse, error)
	StopProvision(provRequest *ProvisionRequest) (*ProvisionResponse, error)
	StartProvision(provRequest *ProvisionRequest) (*ProvisionResponse, error)
	RemoveProvision(provRequest *ProvisionRequest) (*ProvisionResponse, error)
	DeployDag(provRequest *ProvisionRequest) (*ProvisionResponse, error)
	ReProvision(provRequest *ProvisionRequest) (*ProvisionResponse, error)
}

type ProvisionResponse struct {
	OrgID            string        `json:"orgID,omitempty"`
	Hostname         string        `json:"hostName,omitempty"`
	Provisions       []interface{} `json:"provisions,omitempty"`
	ErrorMessage     string        `json:"errorMessage,omitempty"`
	SQLConnectionURL string        `json:"sqlConnectionURL,omitempty"`
	Fernet 			 string 	   `json:"fernet,omitempty"`
	WebserverState   []*TaskState  `json:"webserverState,omitempty"`
	SchedulerState   []*TaskState  `json:"schedulerState,omitempty"`
}

type ProvisionRequest struct {
	OrgID               string                `json:"orgID"`
	Dag                 string                `json:"dag,omitempty"`
	TaskCount           string                `json:"taskCount,omitempty"`
	Parallelism         string                `json:"parallelism,omitempty"`
	WebserverAllocation *ContainerAllocations `json:"webserver,omitempty"`
	SchedulerAllocation *ContainerAllocations `json:"scheduler,omitempty"`
	TaskAllocation      *ContainerAllocations `json:"task,omitempty"`
	MesosRole           string                `json:"mesosRole,omitempty"`
}

type ContainerAllocations struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

type TaskState struct {
	ID     string        `json:"id"`
	Health []*TaskHealth `json:"health"`
	State  string        `json:"state"`
}

type TaskHealth struct {
	State   string `json:"state"`
	Message string `json:"message"`
}
