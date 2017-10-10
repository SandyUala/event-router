# Automated SaaS Deployment Service (ASDS)

## Overview
ASDS is a stateless API that passes calls to various places in order to provision resources for Airflow SaaS deployments. ASDS interfaces with Marathon, Mesos workers directly, the Postgres backend of our Airflow instances and the Houston API. Most of the endpoints are straight forward. A post to `/org/:id` provisions a new postgres user and schema for the organization. It then makes a call through a Marathon client to build the applications. A delete simple undoes this action. A post request with the `:id` neglected means that the `id` will be passed in a json body. A post to the `/dag` endpoint simple sets an environment variable the SaaS Airflow container uses to find the s3 path to pull dags from. Reprovision updates values in Marathon and relaunces the applications. `/resources` and `/metrics` pull different stats from the mesos workers themselves.  `/health` is a function that just returns a 200 response after being called.

## Test Locally
In order to run this locally for testing purposes you'll need to setup ssh tunnels to both the postgres instance with the command `ssh -L 5432:postgres-airflow-astronomer-prod.cki1mapb1yl6.us-east-1.rds.amazonaws.com:5432 core@34.200.223.217 -N` and the mesos instances with the command `ssh -L 8088:marathon.mesos:8080 core@52.55.64.66 -N`. After that you'll need to disable auth on the webserver [here](https://github.com/astronomerio/asds/blob/master/marathon/marathon.go#L166) by adding the line `AddEnv("AIRFLOW__CORE__AUTHENTICATION", "false")` and you'll also need to change the value `c.info.Hostname` [here](https://github.com/astronomerio/asds/blob/master/postgres/postgres.go#L134) to the hostname you used to ssh tunnel into. 

## Reference
**Method**|**Endpoint**|**Go Function**
:-----:|:-----:|:-----:
GET|/org/:id|github.com/astronomerio/asds/api.getOrg
POST|/org|github.com/astronomerio/asds/api.postOrg
POST|/org/:id|github.com/astronomerio/asds/api.postOrg
POST|/org/:id/reprovision|github.com/astronomerio/asds/api.postOrgReprovision
DELETE|/org/:id|github.com/astronomerio/asds/api.deleteOrg
GET|/apps|github.com/astronomerio/asds/api.getApps
POST|/dag|github.com/astronomerio/asds/api.deployDag
GET|/provisions|github.com/astronomerio/asds/api.getProvisions
GET|/resources|github.com/astronomerio/asds/api.getAvailableResources
GET|/metrics|github.com/astronomerio/asds/api.getMetrics
GET|/health|github.com/astronomerio/asds/api.(*Client).Serve.func1

## Example Provision request

### Example Provision Response
```
{
  "orgID": "yourID",
  "hostName": "your-host-name",
  "provisions": [
    {
      "id": "/saas/yourID/airflow/webserver",
      "args": [
        "airflow",
        "webserver"
      ],
      "constraints": [
        [
          "organizationId",
          "LIKE",
          "astronomer"
        ]
      ],
      "container": {
        "type": "DOCKER",
        "docker": {
          "forcePullImage": true,
          "image": "astronomerio/airflow-saas:latest",
          "network": "BRIDGE",
          "parameters": [],
          "portMappings": [
            {
              "containerPort": 8080,
              "hostPort": 0,
              "labels": {},
              "protocol": "tcp"
            }
          ],
          "privileged": false
        },
        "volumes": []
      },
      "cpus": 0.25,
      "gpus": 0,
      "disk": 0,
      "env": {
        "AIRFLOW__CORE__SQL_ALCHEMY_CONN": "postgres://pg_user:pg_pw@pg_hostname:5432/astronomer_saas",
        "AIRFLOW__MESOS__FRAMEWORK_ROLE": "saas",
        "AIRFLOW__MESOS__TASK_PREFIX": "yourID",
        "AIRFLOW__WEBSERVER__AUTHENTICATE": "false",
        "AWS_ACCESS_KEY_ID": "your_access_key_id",
        "AWS_REGION": "us-east-1",
        "AWS_S3_TEMP_BUCKET": "astronomer-workflows",
        "AWS_SECRET_ACCESS_KEY": "your_access_key",
        "S3_ARTIFACT_PATH": ""
      },
      "executor": "",
      "healthChecks": [
        {
          "portIndex": 0,
          "path": "/admin",
          "maxConsecutiveFailures": 3,
          "protocol": "HTTP",
          "gracePeriodSeconds": 30,
          "intervalSeconds": 30,
          "timeoutSeconds": 20,
          "ignoreHttp1xx": false
        }
      ],
      "readinessChecks": [],
      "instances": 1,
      "mem": 512,
      "ports": [
        0
      ],
      "portDefinitions": [
        {
          "port": 0,
          "protocol": "tcp",
          "name": "default",
          "labels": {}
        }
      ],
      "requirePorts": false,
      "backoffSeconds": 1,
      "backoffFactor": 1.15,
      "maxLaunchDelaySeconds": 3600,
      "deployments": [
        {
          "id": "deployment_id"
        }
      ],
      "dependencies": [],
      "upgradeStrategy": {
        "minimumHealthCapacity": 1,
        "maximumOverCapacity": 1
      },
      "unreachableStrategy": {
        "inactiveAfterSeconds": 300,
        "expungeAfterSeconds": 600
      },
      "killSelection": "YOUNGEST_FIRST",
      "uris": [
        "file:///etc/docker.tar.gz"
      ],
      "version": "2017-08-15T20:51:40.929Z",
      "labels": {
        "AIRFLOW_TYPE": "webserver",
        "HAPROXY_0_REDIRECT_TO_HTTPS": "true",
        "HAPROXY_0_VHOST": "your-host.astronomer.io",
        "HAPROXY_GROUP": "external",
        "PROVISIONED_BY": "ASDS",
        "SAAS_HOSTNAME": "your-host",
        "SAAS_ORGANIZATION_ID": "yourID"
      },
      "fetch": [
        {
          "uri": "file:///etc/docker.tar.gz",
          "executable": false,
          "extract": true,
          "cache": false
        }
      ]
    },
    {
      "id": "/saas/yourID/airflow/scheduler",
      "args": [
        "airflow",
        "scheduler",
        "-p"
      ],
      "constraints": [
        [
          "organizationId",
          "LIKE",
          "astronomer"
        ]
      ],
      "container": {
        "type": "DOCKER",
        "docker": {
          "forcePullImage": true,
          "image": "astronomerio/airflow-saas:latest",
          "network": "BRIDGE",
          "parameters": [],
          "portMappings": [],
          "privileged": false
        },
        "volumes": []
      },
      "cpus": 0.25,
      "gpus": 0,
      "disk": 0,
      "env": {
        "AIRFLOW__CORE__EXECUTOR": "astronomer_plugin.AstronomerMesosExecutor",
        "AIRFLOW__CORE__MAX_ACTIVE_RUNS_PER_DAG": "3",
        "AIRFLOW__CORE__SQL_ALCHEMY_CONN": "postgres://pg_user:pg_pw@pg_host:5432/astronomer_saas",
        "AIRFLOW__MESOS__FRAMEWORK_NAME": "airflow",
        "AIRFLOW__MESOS__FRAMEWORK_ROLE": "saas",
        "AIRFLOW__MESOS__MASTER": "leader.mesos:5050",
        "AIRFLOW__MESOS__TASK_PREFIX": "yourID",
        "AWS_ACCESS_KEY_ID": "your_access_key_id",
        "AWS_REGION": "us-east-1",
        "AWS_S3_TEMP_BUCKET": "astronomer-workflows",
        "AWS_SECRET_ACCESS_KEY": "your_secret_key",
        "DOCKER_AIRFLOW_IMAGE_TAG": "astronomerio/airflow-saas:latest",
        "S3_ARTIFACT_PATH": ""
      },
      "executor": "",
      "healthChecks": [
        {
          "command": {
            "value": "ps -aux | grep \"airflow scheduler\" | grep -v grep"
          },
          "maxConsecutiveFailures": 3,
          "protocol": "COMMAND",
          "gracePeriodSeconds": 30,
          "intervalSeconds": 30,
          "timeoutSeconds": 20
        }
      ],
      "readinessChecks": [],
      "instances": 1,
      "mem": 512,
      "ports": [
        0
      ],
      "portDefinitions": [
        {
          "port": 0,
          "protocol": "tcp",
          "name": "default",
          "labels": {}
        }
      ],
      "requirePorts": false,
      "backoffSeconds": 1,
      "backoffFactor": 1.15,
      "maxLaunchDelaySeconds": 3600,
      "deployments": [
        {
          "id": "deployment_id"
        }
      ],
      "dependencies": [],
      "upgradeStrategy": {
        "minimumHealthCapacity": 1,
        "maximumOverCapacity": 1
      },
      "unreachableStrategy": {
        "inactiveAfterSeconds": 300,
        "expungeAfterSeconds": 600
      },
      "killSelection": "YOUNGEST_FIRST",
      "uris": [
        "file:///etc/docker.tar.gz"
      ],
      "version": "2017-08-15T20:51:41.051Z",
      "labels": {
        "AIRFLOW_TYPE": "scheduler",
        "HAPROXY_0_VHOST": "your-host-name.astronomer.io",
        "PROVISIONED_BY": "ASDS",
        "SAAS_HOSTNAME": "your-host-name",
        "SAAS_ORGANIZATION_ID": "yourID"
      },
      "fetch": [
        {
          "uri": "file:///etc/docker.tar.gz",
          "executable": false,
          "extract": true,
          "cache": false
        }
      ]
    }
  ],
  "sqlConnectionURL": "postgres://pg_user:pg_pw@pg_host:5432/astronomer_saas"
  "fernet":"keyhere"
}
```
