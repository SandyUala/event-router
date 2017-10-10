package marathon

import (
	"fmt"
	"strings"

	"github.com/astronomerio/asds/pkg"
	"github.com/spf13/viper"
)

var (
	awsAKIDEnvLabel             = "AWS_ACCESS_KEY_ID"
	awsSAKEnvLabel              = "AWS_SECRET_ACCESS_KEY"
	awsS3TempBucketEnvLabel     = "AWS_S3_TEMP_BUCKET"
	marathonURLEnvLabel         = "MARATHON_URL"
	awsRegionEnvLabel           = "AWS_REGION"
	dbHostNameEnbLabel          = "DB_HOSTNAME"
	dbPortEnvLabel              = "DB_PORT"
	dbUserNameEnvLabel          = "DB_USERNAME"
	dbPasswordEnvLabel          = "DB_PASSWORD"
	dbDatabaseEnvLabel          = "DB_DATABASE"
	airflowImageEnvLabel        = "DOCKER_AIRFLOW_IMAGE_TAG"
	serviceTopLevelNameEnvLabel = "SERVICE_TOP_LEVEL_NAME"
	mesosRoleEnvLabel           = "AIRFLOW__MESOS__FRAMEWORK_ROLE"
	webserverCPUEnvLabel        = "AIRFLOW_WEBSERVER_CPU"
	webserverMemEnvLabel        = "AIRFLOW_WEBSERVER_MEM"
	schedulerCPUEnvLabel        = "AIRFLOW_SCHEDULER_CPU"
	schedulerMemEnvLabel        = "AIRFLOW_SCHEDULER_MEM"
	taskCPUEnvLabel             = "AIRFLOW_TASK_CPU"
	TaskMemEnvLabel             = "AIRFLOW_TASK_MEM"
)

// Config holds the configurations for Marathon
type Config struct {
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	AWSRegion           string
	AWSS3TempBucket     string
	MarathonURL         string
	DBHostname          string
	DBPort              string
	DBUsername          string
	DBPassword          string
	DBDatabase          string
	AirflowImage        string
	ServiceTopLevelName string
	MesosRole           string
	WebserverCPU        float64
	WebserverMem        float64
	SchedulerCPU        float64
	SchedulerMem        float64
	TaskCPU             float64
	TaskMem             float64
}

var config *Config

// GetMarathonConfig returns the configurations for Marathon
func GetMarathonConfig() (*Config, error) {
	if config == nil {
		if err := setupMarathonConfig(); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func setupMarathonConfig() error {
	viper.AutomaticEnv()
	setDefaults()

	var errors []string

	awsAccessKeyID := viper.GetString(awsAKIDEnvLabel)
	if len(awsAccessKeyID) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(awsAKIDEnvLabel))
	}

	awsSecretAccessKey := viper.GetString(awsSAKEnvLabel)
	if len(awsSecretAccessKey) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(awsSAKEnvLabel))
	}

	awsS3TempBucket := viper.GetString(awsS3TempBucketEnvLabel)
	if len(awsS3TempBucket) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(awsS3TempBucketEnvLabel))
	}

	marathonURL := viper.GetString(marathonURLEnvLabel)
	if len(marathonURL) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(marathonURLEnvLabel))
	}

	awsRegion := viper.GetString(awsRegionEnvLabel)
	if len(awsRegion) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(awsRegionEnvLabel))
	}

	dbHostname := viper.GetString(dbHostNameEnbLabel)
	if len(dbHostname) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(dbHostNameEnbLabel))
	}

	dbPort := viper.GetString(dbPortEnvLabel)
	if len(dbPort) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(dbPortEnvLabel))
	}

	dbUsername := viper.GetString(dbUserNameEnvLabel)
	if len(dbUsername) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(dbUserNameEnvLabel))
	}

	dbPassword := viper.GetString(dbPasswordEnvLabel)
	if len(dbPassword) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(dbPasswordEnvLabel))
	}

	dbDatabase := viper.GetString(dbDatabaseEnvLabel)
	if len(dbDatabase) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(dbDatabaseEnvLabel))
	}

	airflowImage := viper.GetString(airflowImageEnvLabel)
	if len(airflowImage) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(airflowImageEnvLabel))
	}

	marathonRole := viper.GetString(mesosRoleEnvLabel)
	if len(marathonRole) == 0 {
		errors = append(errors, pkg.GetRequiredEnvErrorString(mesosRoleEnvLabel))
	}

	serviceTopLevelName := viper.GetString(serviceTopLevelNameEnvLabel)
	webserverCPU := viper.GetFloat64(webserverCPUEnvLabel)
	webserverMem := viper.GetFloat64(webserverMemEnvLabel)
	schedulerCPU := viper.GetFloat64(schedulerCPUEnvLabel)
	schedulerMem := viper.GetFloat64(schedulerMemEnvLabel)
	taskCPU := viper.GetFloat64(taskCPUEnvLabel)
	taskMem := viper.GetFloat64(TaskMemEnvLabel)

	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}

	config = &Config{
		AWSAccessKeyID:      awsAccessKeyID,
		AWSSecretAccessKey:  awsSecretAccessKey,
		AWSRegion:           awsRegion,
		AWSS3TempBucket:     awsS3TempBucket,
		MarathonURL:         marathonURL,
		DBHostname:          dbHostname,
		DBPort:              dbPort,
		DBUsername:          dbUsername,
		DBPassword:          dbPassword,
		DBDatabase:          dbDatabase,
		AirflowImage:        airflowImage,
		ServiceTopLevelName: serviceTopLevelName,
		MesosRole:           marathonRole,
		WebserverCPU:        webserverCPU,
		WebserverMem:        webserverMem,
		SchedulerCPU:        schedulerCPU,
		SchedulerMem:        schedulerMem,
		TaskCPU:             taskCPU,
		TaskMem:             taskMem,
	}

	return nil
}

func setDefaults() {
	viper.SetDefault(serviceTopLevelNameEnvLabel, "saas")
	viper.SetDefault(webserverCPUEnvLabel, 0.25)
	viper.SetDefault(webserverMemEnvLabel, 512)
	viper.SetDefault(schedulerCPUEnvLabel, 0.25)
	viper.SetDefault(schedulerMemEnvLabel, 512)
	viper.SetDefault(taskCPUEnvLabel, 1)
	viper.SetDefault(TaskMemEnvLabel, 3072)

}
