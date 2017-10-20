package clickstream

import (
	"testing"

	"github.com/astronomerio/event-router/s3"
)

var retryBlob = `
{
	"integration":"s3",
	"retryCount":3,
	"message": {
		"messageId":"testID"
	}
}`

func TestS3Upload(t *testing.T) {
	s3Client := &s3.MockClient{}
	p, _ := NewRetryProducer(&ProducerConfig{
		RetryTopic:    "retry",
		RetryS3Bucket: "event-router-messages",
		S3PathPrefix:  "retry",
	}, 2, s3Client)

	message := []byte(retryBlob)
	key := []byte("key")

	p.HandleMessage(message, key)
}
