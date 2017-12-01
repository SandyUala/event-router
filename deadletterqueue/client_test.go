package deadletterqueue

import (
	"testing"

	"github.com/astronomerio/cs-event-router/s3"
)

var (
	key    = "test/key"
	data   = "data"
	bucket = "test"
)

func TestCreate(t *testing.T) {
	s3Client := &s3.MockClient{}
	shutdownChan := make(chan struct{})
	deadletterClient := NewClient(
		&ClientConfig{
			FlushTimeout:    int64(5),
			ShutdownChannel: shutdownChan,
			S3Bucket:        bucket,
			QueueSize:       2,
		}, s3Client)
	deadletterClient.AddToQueue(&QueueObject{
		Key:  key,
		Data: []byte(data),
	})
	close(shutdownChan)
}

func TestFlushMax(t *testing.T) {
	s3Client := &s3.MockClient{}
	shutdownChan := make(chan struct{})
	deadletterClient := NewClient(
		&ClientConfig{
			FlushTimeout:    int64(5),
			ShutdownChannel: shutdownChan,
			S3Bucket:        bucket,
			QueueSize:       2,
		}, s3Client)
	// Add two items then verify it is empty
	qo := &QueueObject{
		Key:  key,
		Data: []byte(data),
	}
	deadletterClient.AddToQueue(qo)
	deadletterClient.AddToQueue(qo)
	if deadletterClient.Length(qo.Key) != 0 {
		log.Error("Deadletter Queue did not flush correctly")
	}
}

func TestFlush(t *testing.T) {
	s3Client := &s3.MockClient{}
	shutdownChan := make(chan struct{})
	deadletterClient := NewClient(
		&ClientConfig{
			FlushTimeout:    int64(5),
			ShutdownChannel: shutdownChan,
			S3Bucket:        bucket,
			QueueSize:       500,
		}, s3Client)
	// Add two items then verify it is empty
	qo := &QueueObject{
		Key:  key,
		Data: []byte(data),
	}
	deadletterClient.AddToQueue(qo)
	deadletterClient.AddToQueue(qo)
	deadletterClient.Flush()
	if deadletterClient.Length(qo.Key) != 0 {
		log.Error("Deadletter Queue did not flush correctly")
	}
}
