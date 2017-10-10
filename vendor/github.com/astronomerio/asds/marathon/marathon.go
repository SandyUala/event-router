package marathon

import (
	"net/url"
	"strconv"
	"time"

	asdsConfig "github.com/astronomerio/asds/config"
	"github.com/astronomerio/asds/pkg"
	"github.com/astronomerio/asds/postgres"
	"github.com/astronomerio/asds/provisioner"
	haikunator "github.com/astronomerio/go-haikunator"
	"github.com/davecgh/go-spew/spew"
	marathon "github.com/gambol99/go-marathon"

	"fmt"

	"database/sql"

	"github.com/sirupsen/logrus"
)

var (
	saasOrganizationID      = "SAAS_ORGANIZATION_ID"
	saasHostName            = "SAAS_HOSTNAME"
	haproxyGroup            = "HAPROXY_GROUP"
	haproxy0RedirectToHTTPS = "HAPROXY_0_REDIRECT_TO_HTTPS"
	haproxy0VHost           = "HAPROXY_0_VHOST"
	provisionedBy           = "PROVISIONED_BY"
	airflowTypeLabel        = "AIRFLOW_TYPE"
	s3ArtifactPathLabel     = "S3_ARTIFACT_PATH"
	sqlConnEnvLabel         = "AIRFLOW__CORE__SQL_ALCHEMY_CONN"
	serviceName             = "%s/%s/airflow/%s"
	taskPrefixEnvLabel      = "AIRFLOW__MESOS__TASK_PREFIX"
	dockerAirflowEnvLabel   = "DOCKER_AIRFLOW_IMAGE_TAG"
)

// Client for the Marathon Provisioner
type Client struct {
	log            *logrus.Entry
	marathonClient marathon.Marathon
	marathonConfig *Config
	dbClient       postgres.PostgresClient
}

type AirflowRequest struct {
	provRqst *provisioner.ProvisionRequest
	HostName string
	SQLConn  string
	Fernet   string
}

// NewProvisioningClient returns a new Marathon Provisioning client
func NewProvisioningClient(log *logrus.Logger, marathonClient marathon.Marathon) (*Client, error) {
	logger := log.WithField("package", "marathon")
	config, err := GetMarathonConfig()
	if err != nil {
		return nil, err
	}
	// Create the SQL Connection info
	dbConfig := &postgres.ConnectionInfo{
		ProvisionerUser: config.DBUsername,
		ProvisionerPw:   config.DBPassword,
		Hostname:        config.DBHostname,
		Port:            config.DBPort,
		DbName:          config.DBDatabase,
	}
	dbClient := postgres.NewClient(dbConfig, log)
	return &Client{
		log:            logger,
		marathonClient: marathonClient,
		marathonConfig: config,
		dbClient:       dbClient,
	}, nil
}

// GetMarathonCLient returns a new Marathon Client
func GetMarathonClient(url string) (marathon.Marathon, error) {
	marathonConfig := marathon.NewDefaultConfig()
	marathonConfig.URL = url
	client, err := marathon.NewClient(marathonConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Provision will provision a new airflow cluster for the given orgID in Marathon
func (c *Client) Provision(provRqst *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	logger := c.log.WithField("function", "Provision")
	logger.WithField("orgConfig", spew.Sdump(provRqst)).Debug("Entered Provision")
	// Check if one already exists
	resp := &provisioner.ProvisionResponse{
		OrgID: provRqst.OrgID,
	}
	exists, hstName, err := c.orgExists(&provRqst.OrgID)
	if err != nil {
		resp.ErrorMessage = err.Error()
		resp.Hostname = hstName
		return resp, err
	}
	if exists {
		err = fmt.Errorf("Origanization %s already exists", provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}

	orgIDHash := pkg.Hash(provRqst.OrgID)
	logger.WithField("hash", orgIDHash).Debug("Calculated Hash from Organization ID")

	// Generate a new hostname from the orgIDHash
	haiku := haikunator.New(orgIDHash)
	hashString := strconv.FormatInt(orgIDHash, 10)
	hostName := fmt.Sprintf("%s-%s", haiku.Haikunate(), hashString[len(hashString)-5:])
	logger.WithField("hostname", hostName).Debug("Generated Host Name")

	db, err := sql.Open("postgres", c.dbClient.GetConnectString())
	if err != nil {
		logger.Fatalf("unable to connect to database %s")
	}
	defer db.Close()

	sqlConn, err := c.dbClient.GetAirflowPostgres(db, provRqst.OrgID)
	if err != nil {
		resp.ErrorMessage = err.Error()
		return resp, err
	}

	c.requestDefaults(provRqst)

	fernet, err := pkg.GetFernet()
	if err != nil {
		logger.Errorf("issue generating fernet key %s", err.Error())
		return resp, err
	}
	resp.Fernet = fernet

	airflowRequest := &AirflowRequest{
		provRqst: provRqst,
		HostName: hostName,
		SQLConn:  sqlConn,
		Fernet:   fernet,
	}
	output := make([]interface{}, 2)
	webserver, err := c.getAirflowWebserver(airflowRequest)
	if err != nil {
		logger.Info("Cleaning DB after failed webserver provision")
		c.cleanUpDB(&provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	webserver, err = c.marathonClient.CreateApplication(webserver)
	if err != nil {
		logger.Info("Cleaning DB after failed webserver provision")
		c.cleanUpDB(&provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	output[0] = webserver

	scheduler, err := c.getAirflowScheduler(airflowRequest)
	if err != nil {
		logger.Info("Cleaning DB after failed webserver provision")
		c.cleanUpDB(&provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	scheduler, err = c.marathonClient.CreateApplication(scheduler)
	if err != nil {
		logger.Info("Cleaning DB after failed webserver provision")
		c.cleanUpDB(&provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	output[1] = scheduler
	resp.Provisions = output
	resp.SQLConnectionURL = sqlConn
	resp.Hostname = hostName
	logger.WithField("Response", spew.Sdump(resp)).Debug("Exiting Provision")
	return resp, nil
}

func (c *Client) getAirflowWebserver(airflowRequest *AirflowRequest) (*marathon.Application, error) {
	logger := c.log.WithField("function", "createAirflowWebserver")
	logger.Debug("Entered createAirflowWebserver")
	if airflowRequest.provRqst.Parallelism == "" {
		airflowRequest.provRqst.Parallelism = "32"
	}
	if airflowRequest.provRqst.TaskCount == "" {
		airflowRequest.provRqst.TaskCount = "16"
	}

	// Create our base Application struct
	app := marathon.NewDockerApplication()

	// Healthcheck Defaults
	maxConsecutiveFailures := 3
	portIndex := 0
	path := "/admin"
	IgnoreHTTP1xx := false

	logsConn := fmt.Sprintf("s3://%s:%s@S3", c.marathonConfig.AWSAccessKeyID, c.marathonConfig.AWSSecretAccessKey)
	logsConnQuery, err := url.ParseQuery(logsConn)
	if err != nil {
		return nil, err
	}
	logsConn = logsConnQuery.Encode()
	logsBucket := "astronomer-customer-airflow"
	logsFolder := fmt.Sprintf("s3://%s/logs/%s", logsBucket, airflowRequest.provRqst.OrgID)
	// Customize our Marathon Application
	app.Name(fmt.Sprintf("/saas/%s/airflow/webserver", airflowRequest.provRqst.OrgID)).
		AddArgs("airflow", "webserver").
		CPU(airflowRequest.provRqst.WebserverAllocation.CPU).
		Memory(airflowRequest.provRqst.WebserverAllocation.Memory).
		Storage(0).
		Count(1).
		AddHealthCheck(marathon.HealthCheck{
			GracePeriodSeconds:     30,
			IntervalSeconds:        30,
			TimeoutSeconds:         20,
			MaxConsecutiveFailures: &maxConsecutiveFailures,
			PortIndex:              &portIndex,
			Path:                   &path,
			Protocol:               "HTTP",
			IgnoreHTTP1xx:          &IgnoreHTTP1xx,
		}).
		AddLabel(haproxy0VHost, fmt.Sprintf("%s.astronomer.io", airflowRequest.HostName)).
		AddLabel(saasOrganizationID, airflowRequest.provRqst.OrgID).
		AddLabel(saasHostName, airflowRequest.HostName).
		AddLabel(provisionedBy, asdsConfig.Prefix).
		AddLabel(airflowTypeLabel, "webserver").
		AddLabel(haproxyGroup, "external").
		AddLabel(haproxy0RedirectToHTTPS, "true").
		AddConstraint("organizationId", "LIKE", "astronomer").
		AddFetchURIs(marathon.Fetch{
			Cache:      false,
			Executable: false,
			Extract:    true,
			URI:        "file:///etc/docker.tar.gz",
		}).
		AddEnv(sqlConnEnvLabel, airflowRequest.SQLConn).
		AddEnv("AIRFLOW__CORE__PARALLELISM", airflowRequest.provRqst.Parallelism).
		AddEnv("AIRFLOW__CORE__DAG_CONCURRENCY", airflowRequest.provRqst.TaskCount).
		AddEnv("AWS_ACCESS_KEY_ID", c.marathonConfig.AWSAccessKeyID).
		AddEnv("AWS_SECRET_ACCESS_KEY", c.marathonConfig.AWSSecretAccessKey).
		AddEnv("AWS_REGION", "us-east-1").
		AddEnv("AWS_S3_TEMP_BUCKET", c.marathonConfig.AWSS3TempBucket).
		AddEnv("ASTRONOMER_ORG_ID", airflowRequest.provRqst.OrgID).
		AddEnv("AIRFLOW_CONN_S3_CONNECTION_LOGS", logsConn).
		AddEnv("AIRFLOW__CORE__REMOTE_BASE_LOG_FOLDER", logsFolder).
		AddEnv("AIRFLOW__CORE__REMOTE_LOG_CONN_ID", "s3_connection_logs").
		AddEnv("AIRFLOW__CORE__FERNET_KEY", airflowRequest.Fernet).
		AddEnv(mesosRoleEnvLabel, airflowRequest.provRqst.MesosRole).
		AddEnv(taskPrefixEnvLabel, airflowRequest.provRqst.OrgID).
		AddEnv(s3ArtifactPathLabel, airflowRequest.provRqst.Dag)

	app.Container.Docker.
		Container(c.marathonConfig.AirflowImage).
		Bridged().
		Expose(8080).
		SetForcePullImage(true)

	logger.Debug("Exiting createAirflowWebserver")
	return app, nil
}

func (c *Client) getAirflowScheduler(airflowRequest *AirflowRequest) (*marathon.Application, error) {
	logger := c.log.WithField("function", "createAirflowScheduler")
	logger.WithField("airflowRequest", spew.Sdump(airflowRequest)).Debug("Entered createAirflowContainer")

	// Create our base Application struct
	app := marathon.NewDockerApplication()

	// Healthcheck Defaults
	maxConsecutiveFailures := 3

	logsConn := fmt.Sprintf("s3://%s:%s@S3", c.marathonConfig.AWSAccessKeyID, c.marathonConfig.AWSSecretAccessKey)
	logsConnQuery, err := url.ParseQuery(logsConn)
	if err != nil {
		return nil, err
	}
	logsConn = logsConnQuery.Encode()

	logsBucket := "astronomer-customer-airflow"
	logsFolder := fmt.Sprintf("s3://%s/logs/%s", logsBucket, airflowRequest.provRqst.OrgID)
	taskCPU := strconv.FormatFloat(airflowRequest.provRqst.TaskAllocation.CPU, 'f', -1, 64)
	taskMemory := strconv.FormatFloat(airflowRequest.provRqst.TaskAllocation.Memory, 'f', -1, 64)
	// Customize our Marathon Application
	app.Name(fmt.Sprintf("/saas/%s/airflow/scheduler", airflowRequest.provRqst.OrgID)).
		AddArgs("airflow", "scheduler").
		CPU(airflowRequest.provRqst.SchedulerAllocation.CPU).
		Memory(airflowRequest.provRqst.SchedulerAllocation.Memory).
		Storage(0).
		Count(1).
		AddHealthCheck(marathon.HealthCheck{
			GracePeriodSeconds:     30,
			IntervalSeconds:        30,
			TimeoutSeconds:         20,
			MaxConsecutiveFailures: &maxConsecutiveFailures,
			Command: &marathon.Command{
				Value: "ps -aux | grep \"airflow scheduler\" | grep -v grep",
			},
			Protocol: "COMMAND",
		}).
		AddLabel(haproxy0VHost, fmt.Sprintf("%s.astronomer.io", airflowRequest.HostName)).
		AddLabel(saasOrganizationID, airflowRequest.provRqst.OrgID).
		AddLabel(saasHostName, airflowRequest.HostName).
		AddLabel(provisionedBy, asdsConfig.Prefix).
		AddLabel(airflowTypeLabel, "scheduler").
		AddConstraint("organizationId", "LIKE", "astronomer").
		AddFetchURIs(marathon.Fetch{
			Cache:      false,
			Executable: false,
			Extract:    true,
			URI:        "file:///etc/docker.tar.gz",
		}).
		AddEnv(sqlConnEnvLabel, airflowRequest.SQLConn).
		AddEnv("AIRFLOW__CORE__MAX_ACTIVE_RUNS_PER_DAG", "3").
		AddEnv("AIRFLOW__MESOS__MASTER", "leader.mesos:5050").
		AddEnv("AIRFLOW__CORE__EXECUTOR", "astronomer_plugin.AstronomerMesosExecutor").
		AddEnv("AIRFLOW__MESOS__FRAMEWORK_NAME", "airflow").
		AddEnv(dockerAirflowEnvLabel, c.marathonConfig.AirflowImage).
		AddEnv("AWS_ACCESS_KEY_ID", c.marathonConfig.AWSAccessKeyID).
		AddEnv("AWS_SECRET_ACCESS_KEY", c.marathonConfig.AWSSecretAccessKey).
		AddEnv("AWS_REGION", "us-east-1").
		AddEnv("AWS_S3_TEMP_BUCKET", c.marathonConfig.AWSS3TempBucket).
		AddEnv("ASTRONOMER_ORG_ID", airflowRequest.provRqst.OrgID).
		AddEnv("AIRFLOW_CONN_S3_CONNECTION_LOGS", logsConn).
		AddEnv("AIRFLOW__CORE__REMOTE_BASE_LOG_FOLDER", logsFolder).
		AddEnv("AIRFLOW__CORE__REMOTE_LOG_CONN_ID", "s3_connection_logs").
		AddEnv("AIRFLOW__CORE__FERNET_KEY", airflowRequest.Fernet).
		AddEnv("AIRFLOW__MESOS__TASK_CPU", taskCPU).
		AddEnv("AIRFLOW__MESOS__TASK_MEMORY", taskMemory).
		AddEnv(mesosRoleEnvLabel, airflowRequest.provRqst.MesosRole).
		AddEnv(taskPrefixEnvLabel, airflowRequest.provRqst.OrgID).
		AddEnv(s3ArtifactPathLabel, airflowRequest.provRqst.Dag)

	app.Container.Docker.
		Container(c.marathonConfig.AirflowImage).
		Bridged().
		SetForcePullImage(true)

	logger.Debug("Exiting createAirflowWebserver")
	return app, nil
}

func (c *Client) requestDefaults(provRqst *provisioner.ProvisionRequest) {
	// Set container allocations to default if they are not specified
	if provRqst.WebserverAllocation == nil {
		provRqst.WebserverAllocation = &provisioner.ContainerAllocations{
			CPU:    c.marathonConfig.WebserverCPU,
			Memory: c.marathonConfig.WebserverMem,
		}
	} else {
		if provRqst.WebserverAllocation.CPU == 0 {
			provRqst.WebserverAllocation.CPU = c.marathonConfig.WebserverCPU
		}
		if provRqst.WebserverAllocation.Memory == 0 {
			provRqst.WebserverAllocation.Memory = c.marathonConfig.WebserverMem
		}
	}
	if provRqst.SchedulerAllocation == nil {
		provRqst.SchedulerAllocation = &provisioner.ContainerAllocations{
			CPU:    c.marathonConfig.SchedulerCPU,
			Memory: c.marathonConfig.SchedulerMem,
		}
	} else {
		if provRqst.SchedulerAllocation.CPU == 0 {
			provRqst.SchedulerAllocation.CPU = c.marathonConfig.SchedulerCPU
		}
		if provRqst.SchedulerAllocation.Memory == 0 {
			provRqst.SchedulerAllocation.Memory = c.marathonConfig.SchedulerMem
		}
	}
	if provRqst.TaskAllocation == nil {
		provRqst.TaskAllocation = &provisioner.ContainerAllocations{
			CPU:    c.marathonConfig.TaskCPU,
			Memory: c.marathonConfig.TaskMem,
		}
	} else {
		if provRqst.TaskAllocation.CPU == 0 {
			provRqst.TaskAllocation.CPU = c.marathonConfig.TaskCPU
		}
		if provRqst.TaskAllocation.Memory == 0 {
			provRqst.TaskAllocation.Memory = c.marathonConfig.TaskMem
		}
	}

	if len(provRqst.MesosRole) == 0 {
		provRqst.MesosRole = c.marathonConfig.MesosRole
	}
}

// ListProvisions will list all the provisions in Marathon
func (c *Client) ListProvisions() (*provisioner.ProvisionResponse, error) {
	logger := c.log.WithField("function", "ListProvision")
	logger.Debug("Entered ListProvisions")
	resp := &provisioner.ProvisionResponse{}
	v := url.Values{}
	v["label"] = []string{provisionedBy + "==" + asdsConfig.Prefix}
	logger.WithField("URL Values", v.Encode()).Debug("Setting URL Values")
	apps, err := c.marathonClient.Applications(v)
	if err != nil {
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	if apps == nil || len(apps.Apps) == 0 {
		// If we found no provisions, return nil
		return nil, nil
	}
	var provisions []*provisioner.ProvisionResponse
	requestedOrgIDs := make(map[string]int)
	for _, app := range apps.Apps {
		orgID := (*app.Labels)[saasOrganizationID]
		if _, ok := requestedOrgIDs[orgID]; !ok {
			// Get the provision information
			requestedOrgIDs[orgID] = 1
			prov, err := c.GetProvision(&provisioner.ProvisionRequest{OrgID: orgID})
			if err != nil {
				prov.OrgID = orgID
				prov.ErrorMessage = err.Error()
			}
			provisions = append(provisions, prov)
		}
	}
	var output []interface{}
	resp.Provisions = append(output, provisions)
	return resp, nil
}

// GetProvision will return the provision information for the given orginization from Marathon
func (c *Client) GetProvision(provRqst *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	logger := c.log.WithField("funciton", "GetProvision")
	logger.WithField("orgID", provRqst.OrgID).Debug("Entered GetProvision")
	v := url.Values{}
	v["label"] = []string{saasOrganizationID + "==" + provRqst.OrgID}
	apps, err := c.marathonClient.Applications(v)
	if err != nil {
		logger.Debug("getApps failed")
		return &provisioner.ProvisionResponse{ErrorMessage: err.Error()}, err
	}
	if apps == nil || len(apps.Apps) == 0 {
		// If we found no provisions, return nil
		return nil, nil
	}
	hostname := ""
	sqlConn := ""
	var output []interface{}
	resp := &provisioner.ProvisionResponse{
		OrgID:    provRqst.OrgID,
		Hostname: hostname,
	}
	for _, app := range apps.Apps {
		output = append(output, app)
		hostname = (*app.Labels)[saasHostName]
		sqlConn = (*app.Env)["sqlConnEnvLabel"]
		// Get Task State
		tasks, err := c.marathonClient.Tasks(app.ID)
		if err != nil {
			resp.ErrorMessage = err.Error()
			return resp, err
		}
		for _, task := range tasks.Tasks {
			taskState := &provisioner.TaskState{}
			if task.HasHealthCheckResults() {
				var lastFailure time.Time
				var lastSuccess time.Time
				for _, healthCheck := range task.HealthCheckResults {
					haveFailure, haveSuccess := false, false
					if healthCheck.LastFailure != "" {
						lastFailure, err = time.Parse(time.RFC3339, healthCheck.LastFailure)
						haveFailure = true
						if err != nil {
							resp.ErrorMessage = err.Error()
							return resp, err
						}
					}
					if healthCheck.LastSuccess != "" {
						lastSuccess, err = time.Parse(time.RFC3339, healthCheck.LastSuccess)
						haveSuccess = true
						if err != nil {
							resp.ErrorMessage = err.Error()
							return resp, err
						}
					}
					if haveFailure && !haveSuccess {
						taskState.Health = append(taskState.Health, &provisioner.TaskHealth{
							State:   "unhealthy",
							Message: healthCheck.LastFailureCause,
						})
					} else if haveSuccess && !haveFailure {
						taskState.Health = append(taskState.Health, &provisioner.TaskHealth{
							State: "healthy",
						})
					} else if haveSuccess && haveFailure {
						if lastFailure.After(lastSuccess) {
							taskState.Health = append(taskState.Health, &provisioner.TaskHealth{
								State:   "unhealthy",
								Message: healthCheck.LastFailureCause,
							})
						} else {
							taskState.Health = append(taskState.Health, &provisioner.TaskHealth{
								State: "healthy",
							})
						}
					} else {
						taskState.Health = append(taskState.Health, &provisioner.TaskHealth{
							State:   "unhealthy",
							Message: healthCheck.LastFailureCause,
						})
					}
				}
			}
			taskState.ID = task.ID
			taskState.State = task.State
			switch (*app.Labels)[airflowTypeLabel] {
			case "scheduler":
				resp.SchedulerState = append(resp.SchedulerState, taskState)
			case "webserver":
				resp.WebserverState = append(resp.WebserverState, taskState)
			}
		}
	}
	resp.Provisions = output
	resp.SQLConnectionURL = sqlConn
	logger.Debug("Exiting GetProvision")
	return resp, nil
}

func (c *Client) StopProvision(provRequest *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	return nil, nil
}

func (c *Client) StartProvision(provRequest *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	return nil, nil
}

func (c *Client) RemoveProvision(provRqst *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	logger := c.log.WithField("function", "RemoveProvision")
	logger.WithField("orgID", provRqst.OrgID).Debug("Entered RemoveProvision")
	// Destroy the scheduler then webserver
	resp := &provisioner.ProvisionResponse{}
	resp.OrgID = provRqst.OrgID

	exists, _, err := c.orgExists(&provRqst.OrgID)
	if err != nil {
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	if exists {
		_, err = c.marathonClient.DeleteApplication(fmt.Sprintf(serviceName, c.marathonConfig.ServiceTopLevelName, provRqst.OrgID, "webserver"), true)
		if err != nil {
			resp.ErrorMessage = err.Error()
			return resp, err
		}

		_, err = c.marathonClient.DeleteApplication(fmt.Sprintf(serviceName, c.marathonConfig.ServiceTopLevelName, provRqst.OrgID, "scheduler"), true)
		if err != nil {
			resp.ErrorMessage = err.Error()
			return resp, err
		}

		// Remove left over group
		_, err = c.marathonClient.DeleteGroup(fmt.Sprintf("%s/%s", c.marathonConfig.ServiceTopLevelName, provRqst.OrgID), true)
		if err != nil {
			resp.ErrorMessage = err.Error()
			return resp, err
		}
	}

	db, err := sql.Open("postgres", c.dbClient.GetConnectString())
	if err != nil {
		logger.Fatalf("unable to open database connection %s", err.Error())
	}
	defer db.Close()

	if err = c.dbClient.DestroyAirflowPostgres(db, provRqst.OrgID); err != nil {
		logger.WithError(err).Debug("Error dropping user or schema")
		resp.ErrorMessage = err.Error()
		return resp, err
	}

	logger.Debug("Exiting Remove Provision")

	return nil, nil
}

// DeployDag makes a ReProvision request with updated dag information.
func (c *Client) DeployDag(provRqst *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	// In the end we reprovision with updated dag env, just call that function
	return c.ReProvision(provRqst)
}

// ReProvision updates an existing deployment with new request values.
func (c *Client) ReProvision(provRqst *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	logger := c.log.WithField("function", "ReProvision")
	logger.WithFields(logrus.Fields{"orgID": provRqst.OrgID, "dagPath": provRqst.Dag}).Debug("Entered ReProvision")
	// Get the webserver app

	resp := &provisioner.ProvisionResponse{
		OrgID: provRqst.OrgID,
	}

	webserver, err := c.marathonClient.Application(fmt.Sprintf(serviceName, c.marathonConfig.ServiceTopLevelName, provRqst.OrgID, "webserver"))
	if err != nil {
		return nil, err
	}
	if webserver == nil {
		err = fmt.Errorf("Aiflow servers not found for %s", provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}

	scheduler, err := c.marathonClient.Application(fmt.Sprintf(serviceName, c.marathonConfig.ServiceTopLevelName, provRqst.OrgID, "scheduler"))
	if err != nil {
		return nil, err
	}
	if scheduler == nil {
		err = fmt.Errorf("Aiflow servers not found for %s", provRqst.OrgID)
		resp.ErrorMessage = err.Error()
		return resp, err
	}

	// Check out reprovision request.  If it doesn't have the dag specified, use the old dag
	if len(provRqst.Dag) == 0 {
		provRqst.Dag = (*scheduler.Env)[s3ArtifactPathLabel]
	}

	if provRqst.SchedulerAllocation == nil {
		provRqst.SchedulerAllocation = &provisioner.ContainerAllocations{
			CPU:    scheduler.CPUs,
			Memory: *scheduler.Mem,
		}
	}
	if provRqst.WebserverAllocation == nil {
		provRqst.WebserverAllocation = &provisioner.ContainerAllocations{
			CPU:    webserver.CPUs,
			Memory: *webserver.Mem,
		}
	}

	webserverRequest := &AirflowRequest{
		provRqst: provRqst,
		HostName: (*webserver.Labels)[saasHostName],
		SQLConn:  (*webserver.Env)[sqlConnEnvLabel],
	}

	webserver, err = c.getAirflowWebserver(webserverRequest)
	if err != nil {
		return nil, err
	}

	schedulerRequest := &AirflowRequest{
		provRqst: provRqst,
		HostName: (*scheduler.Labels)[saasHostName],
		SQLConn:  (*scheduler.Env)[sqlConnEnvLabel],
	}

	scheduler, err = c.getAirflowScheduler(schedulerRequest)
	if err != nil {
		return nil, err
	}

	var output []interface{}

	// Update webserver
	dep, err := c.marathonClient.UpdateApplication(webserver, true)
	if err != nil {
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	output = append(output, dep)

	// Update scheduler
	dep, err = c.marathonClient.UpdateApplication(scheduler, true)
	if err != nil {
		resp.ErrorMessage = err.Error()
		return resp, err
	}
	output = append(output, dep)

	resp.Provisions = output
	return resp, nil
}

func (c *Client) orgExists(orgID *string) (bool, string, error) {
	logger := c.log.WithField("function", "orgExists")
	logger.Debug("Entered orgExists")
	app, err := c.GetProvision(&provisioner.ProvisionRequest{OrgID: *orgID})
	logger.WithField("app", spew.Sdump(app)).Debug("app info")
	if err != nil {
		return false, "", err
	}
	hostName := ""
	if app != nil {
		hostName = app.Hostname
	}
	return app != nil, hostName, err
}

// GetAppApplications will return all applications currently running on marathon.
// This is mainly used for internal testing.
func (c *Client) GetAllApplications() (*provisioner.ProvisionResponse, error) {
	var output []interface{} // Generic interface so we can just print it out
	apps, err := c.marathonClient.Applications(nil)
	if err != nil {
		return nil, err
	}
	for _, app := range apps.Apps {
		details, err := c.marathonClient.Application(app.ID)
		if err != nil {
			return nil, err
		}
		output = append(output, *details)
	}
	prov := &provisioner.ProvisionResponse{}
	prov.Provisions = output
	return prov, nil
}

func (c *Client) cleanUpDB(orgID *string) error {
	logger := c.log.WithField("function", "cleanUpDB")
	logger.Debug("Entered cleanUpDB")
	db, err := sql.Open("postgres", c.dbClient.GetConnectString())
	if err != nil {
		logger.Fatalf("unable to connect to database %s", err.Error())
	}
	defer db.Close()

	return c.dbClient.DestroyAirflowPostgres(db, *orgID)
}
