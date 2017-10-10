package marathon

import (
	"net/url"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

type Marathon struct {
}

func NewTestMarathonclient() marathon.Marathon {
	return Marathon{}
}

// get a listing of the application ids
func (m Marathon) ListApplications(url.Values) ([]string, error) {
	return nil, nil
}

// a list of application versions
func (m Marathon) ApplicationVersions(name string) (*marathon.ApplicationVersions, error) {
	return nil, nil
}

// check a application version exists
func (m Marathon) HasApplicationVersion(name, version string) (bool, error) {
	return false, nil
}

// change an application to a different version
func (m Marathon) SetApplicationVersion(name string, version *marathon.ApplicationVersion) (*marathon.DeploymentID, error) {
	return nil, nil
}

// check if an application is ok
func (m Marathon) ApplicationOK(name string) (bool, error) {
	return false, nil
}

// create an application in marathon
func (m Marathon) CreateApplication(application *marathon.Application) (*marathon.Application, error) {
	// Return the app passed in with no error
	return application, nil
}

// delete an application
func (m Marathon) DeleteApplication(name string, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// update an application in marathon
func (m Marathon) UpdateApplication(application *marathon.Application, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// a list of deployments on a application
func (m Marathon) ApplicationDeployments(name string) ([]*marathon.DeploymentID, error) {
	return nil, nil
}

// scale a application
func (m Marathon) ScaleApplicationInstances(name string, instances int, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// restart an application
func (m Marathon) RestartApplication(name string, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// get a list of applications from marathon
func (m Marathon) Applications(url.Values) (*marathon.Applications, error) {
	return &marathon.Applications{}, nil
}

// get an application by name
func (m Marathon) Application(name string) (*marathon.Application, error) {
	return nil, nil
}

// get an application by options
func (m Marathon) ApplicationBy(name string, opts *marathon.GetAppOpts) (*marathon.Application, error) {
	return nil, nil
}

// get an application by name and version
func (m Marathon) ApplicationByVersion(name, version string) (*marathon.Application, error) {
	return nil, nil
}

// wait of application
func (m Marathon) WaitOnApplication(name string, timeout time.Duration) error {
	return nil
}

// -- TASKS ---

// get a list of tasks for a specific application
func (m Marathon) Tasks(application string) (*marathon.Tasks, error) {
	return nil, nil
}

// get a list of all tasks
func (m Marathon) AllTasks(opts *marathon.AllTasksOpts) (*marathon.Tasks, error) {
	return nil, nil
}

// get the endpoints for a service on a application
func (m Marathon) TaskEndpoints(name string, port int, healthCheck bool) ([]string, error) {
	return nil, nil
}

// kill all the tasks for any application
func (m Marathon) KillApplicationTasks(applicationID string, opts *marathon.KillApplicationTasksOpts) (*marathon.Tasks, error) {
	return nil, nil
}

// kill a single task
func (m Marathon) KillTask(taskID string, opts *marathon.KillTaskOpts) (*marathon.Task, error) {
	return nil, nil
}

// kill the given array of tasks
func (m Marathon) KillTasks(taskIDs []string, opts *marathon.KillTaskOpts) error {
	return nil
}

// --- GROUPS ---

// list all the groups in the system
func (m Marathon) Groups() (*marathon.Groups, error) {
	return nil, nil
}

// retrieve a specific group from marathon
func (m Marathon) Group(name string) (*marathon.Group, error) {
	return nil, nil
}

// list all groups in marathon by options
func (m Marathon) GroupsBy(opts *marathon.GetGroupOpts) (*marathon.Groups, error) {
	return nil, nil
}

// retrieve a specific group from marathon by options
func (m Marathon) GroupBy(name string, opts *marathon.GetGroupOpts) (*marathon.Group, error) {
	return nil, nil
}

// create a group deployment
func (m Marathon) CreateGroup(group *marathon.Group) error {
	return nil
}

// delete a group
func (m Marathon) DeleteGroup(name string, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// update a groups
func (m Marathon) UpdateGroup(id string, group *marathon.Group, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// check if a group exists
func (m Marathon) HasGroup(name string) (bool, error) {
	return false, nil
}

// wait for an group to be deployed
func (m Marathon) WaitOnGroup(name string, timeout time.Duration) error {
	return nil
}

// --- DEPLOYMENTS ---

// get a list of the deployments
func (m Marathon) Deployments() ([]*marathon.Deployment, error) {
	return nil, nil
}

// delete a deployment
func (m Marathon) DeleteDeployment(id string, force bool) (*marathon.DeploymentID, error) {
	return nil, nil
}

// check to see if a deployment exists
func (m Marathon) HasDeployment(id string) (bool, error) {
	return false, nil
}

// wait of a deployment to finish
func (m Marathon) WaitOnDeployment(id string, timeout time.Duration) error {
	return nil
}

// --- SUBSCRIPTIONS ---

// a list of current subscriptions
func (m Marathon) Subscriptions() (*marathon.Subscriptions, error) {
	return nil, nil
}

// add a events listener
func (m Marathon) AddEventsListener(filter int) (marathon.EventsChannel, error) {
	return nil, nil
}

// remove a events listener
func (m Marathon) RemoveEventsListener(channel marathon.EventsChannel) {

}

// Subscribe a callback URL
func (m Marathon) Subscribe(string) error {
	return nil
}

// Unsubscribe a callback URL
func (m Marathon) Unsubscribe(string) error {
	return nil
}

// --- QUEUE ---
// get marathon launch queue
func (m Marathon) Queue() (*marathon.Queue, error) {
	return nil, nil
}

// resets task launch delay of the specific application
func (m Marathon) DeleteQueueDelay(appID string) error {
	return nil
}

// --- MISC ---

// get the marathon url
func (m Marathon) GetMarathonURL() string {
	return ""
}

// ping the marathon
func (m Marathon) Ping() (bool, error) {
	return false, nil
}

// grab the marathon server info
func (m Marathon) Info() (*marathon.Info, error) {
	return nil, nil
}

// retrieve the leader info
func (m Marathon) Leader() (string, error) {
	return "", nil
}

// cause the current leader to abdicate
func (m Marathon) AbdicateLeader() (string, error) {
	return "", nil
}
