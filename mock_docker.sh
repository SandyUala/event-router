#!/bin/bash

export ER_BOOTSTRAP_SERVERS=localhost:9092
export ER_HOUSTON_API_URL=localhost
export ER_HOUSTON_API_KEY=key
export ER_DEBUG=true
export ER_SERVE_PORT=9091
export ER_TOPIC=mocktopic
export ER_GROUP_ID=moc
export ER_SSE_URL=https://houston.astronomer.io/broadcast
export ER_SSE_AUTH="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjU5OWFmMzlhMDg5N2E3MDAwMWVlN2VlNSIsInNVIjp0cnVlLCJpYXQiOjE1MDc5MTM4ODAsImV4cCI6MTUwODUxODY4MH0.8d76ZYaNdQliCFtUZYgE31iXVA3hDB9MePvaXX3pzp0"


docker run -it --rm --net=host \
  -e ER_BOOTSTRAP_SERVERS=$ER_BOOTSTRAP_SERVERS \
  -e ER_HOUSTON_API_URL=$ER_HOUSTON_API_URL \
  -e ER_HOUSTON_API_KEY=$ER_HOUSTON_API_KEY \
  -e ER_DEBUG=$ER_DEBUG \
  -e ER_SERVE_PORT=$ER_SERVE_PORT \
  -e ER_TOPIC=$ER_TOPIC \
  -e ER_GROUP_ID=$ER_GROUP_ID \
  -e ER_SSE_URL=$ER_SSE_URL\
  -e ER_SSE_AUTH=$ER_SSE_AUTH \
  astronomerio/event-router mock "S3 Event Logs:s3-event-logs"