# Clickstream Event Router

Clickstream Event Router is responsible for routing clickstream events to the appropriate integrations based on application ID.  It consumes messages from a kafka topic, then retrieves the enabled integrations from Houston, and produces the same message to the integrations kafka topic.

It has its own internal cache of enabled integrations for each application ID.  If the cache does not have an entry for an app ID, it retrieves the enabled integrations from Houston.  Event Router also subscribes to a broadcast channel from Houston to get integration change events.  When an event is received by Event Router, it will update the app IDs integrations with the latest from Houston.

~~The cache by default has a 5 min TTL that can be changed or disabled.  This ensures that the cache will always have up to date information in case a broadcast message is missed.~~ Not merged in yet.

### Usage:

```
event-router will route incoming events from analytics.js to the correct integration

Usage:
  event-router [command]

Available Commands:
  help        Help about any command
  mock        run even-router with a mock houston, takes a list of integrations to enabled
  start       start event-router

Flags:
      --disable-sse   disables SSE client
  -h, --help          help for event-router

Use "event-router [command] --help" for more information about a command.
```

### Environment Variables

 * `ER_DEBUG`
   * Optional, Default: false
   * Enables debug logging
 * `ER_BOOTSTRAP_SERVERS`
   * Required
   * Format: IP/URL with Port, separated by comma (`127.0.0.1:9090,127.0.1.1:9090`)
 * `ER_SERVER_PORT`
   * Optional, Default: `8080`
 * `ER_KAFKA_GROUP_ID`
   * Required. Kafka uses Group ID to group apps together so each one gets different messages.
   * Format: String
 * `ER_KAFKA_INGESTION_TOPIC`
   * Required.  Main ingestion topic
 * `ER_SSE_URL`
   * Required.  URL to the SSE Broadcast channel (usually Houstons api endpoint `/broadcast`)
 * `ER_KAFKA_PRODUCER_FLUSH_TIMEOUT_MS`
   * Optional, Default: `1000`
   * Timeout when flushing messages to kafka.  Used when shutting down the event-router.  If any messages are left after flushing, they are dropped.
 * `ER_KAFKA_PRODUCER_MESSAGE_TIMOUT_MS`
   * Optional, Default: `5000`
   * Message Timeout for Kafka
 * `ER_MAX_RETRIES`
   * Optional, Default: `2`
   * Number of times the producer will retry a failed message.  Does not take into account librdkafka's built in retry functionality (which is 2 by default)
 * `ER_CLICKSTREAM_RETRY_TOPIC`
   * Required if RETRY is enabled
   * Kafka Topic that retry messages are produced too
 * `ER_CLICKSTREAM_RETRY_S3_BUCKET`
   * Required if RETRY is enabled
   * S3 Bucket messages are sent to if they are not successfully sent after retrying.  File name is the `messageId` and contents is json.
 * `CLICKSTREAM_RETRY_S3_PATH_PREFIX`
   * Optional
   * Path Prefix used when saving the message to S3
 * `ER_HOUSTON_API_URL`
   * Required
   * Houston API URL
 * `ER_HOUSTON_API_KEY`
   * Required if the Houston Username/Password is not specified
   * API Key to access Houston.  Requires `superuser` level access
 * `ER_HOUSTON_USERNAME`
   * Required if `ER_HOUSTON_API_KEY` is not specified
   * Username to Houston user with `superuser` level access
 * `ER_HOUSTON_PASSWORD`
   * Required if `ER_HOUSTON_API_KEY` is not specified
   * Password to Houston user account