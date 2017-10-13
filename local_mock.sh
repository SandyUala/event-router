#!/bin/bash

export ER_BOOTSTRAP_SERVERS=localhost:9092
export ER_HOUSTON_API_URL=localhost
export ER_HOUSTON_API_KEY=key
export ER_DEBUG=TRUE
export ER_SERVE_PORT=9091
export ER_TOPIC=main
export ER_GROUP_ID=moc
export ER_SSE_URL=https://houston.astronomer.io/broadcast
export ER_SSE_AUTH="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjU5OWFmMzlhMDg5N2E3MDAwMWVlN2VlNSIsInNVIjp0cnVlLCJpYXQiOjE1MDc5MTM4ODAsImV4cCI6MTUwODUxODY4MH0.8d76ZYaNdQliCFtUZYgE31iXVA3hDB9MePvaXX3pzp0"

go build
./event-router mock "S3 Event Logs:s3-event-logs"
