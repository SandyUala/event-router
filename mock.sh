#!/bin/bash

export ER_BOOTSTRAP_SERVERS=localhost:9092
export ER_HOUSTON_API_URL=localhost
export ER_HOUSTON_API_KEY=key
export ER_DEBUG=false
export ER_SERVE_PORT=9091
export ER_TOPIC=main
export ER_GROUP_ID=moc

go run main.go mock "S3 Event Logs:s3-event-logs"