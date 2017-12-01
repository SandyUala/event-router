package deadletterqueue

import (
	"time"

	"sync"

	"strings"

	"bytes"

	"strconv"

	"github.com/astronomerio/cs-event-router/pkg/prom"
	"github.com/astronomerio/cs-event-router/s3"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// DeadLetterQueue Interface
type DeadLetterQueue interface {
	Start()
	AddToQueue(qo *QueueObject) error
	Flush() error
	Close()
}

// QueueObject contains the Key and Data to be added to the DeadLetter Queue
// The key must be in the format "appID/Integration" as it is used to save the
// file to S3
type QueueObject struct {
	Data []byte
	Key  string
}

// Client is a DeadLetter Queue client.  It holds the timer, s3client, and queue object.
type Client struct {
	timer           *time.Timer
	queueSize       int64
	s3Client        s3.S3Client
	queue           map[string][][]byte
	globalLock      *sync.Mutex
	s3Bucket        string
	shutdownChannel chan struct{}
	addingWaitGroup *sync.WaitGroup
}

// ClientConfig contains configuration parameters for a DeadLetter Client
type ClientConfig struct {
	FlushTimeout    int64
	QueueSize       int64
	S3Bucket        string
	ShutdownChannel chan struct{}
}

var (
	log = logrus.WithField("package", "deadletterqueue")
)

// NewClient returns a new DeadLetter Queue Client
func NewClient(cfg *ClientConfig, s3Client s3.S3Client) *Client {
	client := &Client{
		timer:           time.NewTimer(time.Minute * time.Duration(cfg.FlushTimeout)),
		queueSize:       cfg.QueueSize,
		s3Client:        s3Client,
		globalLock:      &sync.Mutex{},
		queue:           make(map[string][][]byte),
		s3Bucket:        cfg.S3Bucket,
		shutdownChannel: cfg.ShutdownChannel,
		addingWaitGroup: &sync.WaitGroup{},
	}
	go client.start()
	return client
}

// AddToQueue adds the QueueObject to the internal queue.  If the queue for the given Key
// reaches the Queue Size, it will be flushed to S3
func (c *Client) AddToQueue(qo *QueueObject) error {
	logger := log.WithField("function", "AddToQueue")
	logger.WithField("key", qo.Key).Debug("adding to deadletter queue")
	if len(qo.Key) == 0 {
		return errors.New("Dead Letter Queue key cannot be empty")
	}
	// If there is no data don't bother saving it
	if len(qo.Data) == 0 {
		return nil
	}
	if !strings.Contains(qo.Key, "/") {
		return errors.New("invalid key, must be in the format 'appID/Integration'")
	}
	// Get a lock on the map so we can update it
	c.globalLock.Lock()
	c.addingWaitGroup.Add(1)
	defer c.addingWaitGroup.Done()
	defer c.globalLock.Unlock()
	// Add our new data to the queue
	c.queue[qo.Key] = append(c.queue[qo.Key], qo.Data)
	// If we are over the Queue Size, flush it.
	if int64(len(c.queue[qo.Key])) >= c.queueSize {
		if err := c.flush(qo.Key); err != nil {
			return errors.Wrap(err, "error flushing dead letter queue after hitting queue size")
		}
	}

	return nil
}

func (c *Client) start() {
	logger := log.WithField("function", "start")
	logger.Info("Starting Deadletter Queue")
	// start the timer
	for {
		select {
		case <-c.timer.C:
			if err := c.Flush(); err != nil {
				logger.Error(err)
			}
		case <-c.shutdownChannel:
			logger.Info("Deadletter Queue shutting down")
			c.Close()
			return
		}
	}
}

// Flush will flush the content of the queue to S3
func (c *Client) Flush() error {
	logger := log.WithField("function", "Flush")
	// Lock the queue so we can flush it out
	c.globalLock.Lock()
	defer c.globalLock.Unlock()
	errMessages := []string{}
	for key, _ := range c.queue {
		if err := c.flush(key); err != nil {
			logger.WithField("error", err).Error("Error flushing data")
			errMessages = append(errMessages, err.Error())
		}
	}
	if len(errMessages) != 0 {
		return errors.New(strings.Join(errMessages, ", "))
	}
	return nil
}

func (c *Client) flush(key string) error {
	logger := log.WithField("function", "flush")
	logger.WithField("key", key).Debug("Flushing Deadletter Queue")
	// create a buffer to add all the messages too
	data := []byte{}
	buffer := bytes.NewBuffer(data)
	var count, size float64
	for _, value := range c.queue[key] {
		s, _ := buffer.Write(value)
		size += float64(s)
		s, _ = buffer.WriteString("\n")
		size += float64(s)
		count++
	}
	now := strconv.FormatInt(time.Now().Unix(), 10)
	key = key + "/" + now
	if err := c.s3Client.SendToS3(&c.s3Bucket, &key, data); err != nil {
		return errors.Wrap(err, "error flushing "+key)
	}
	// Update prometheus
	values := strings.Split(key, "/")
	prom.MessagesUploadedToS3.With(prometheus.Labels{"appId": values[0], "integration": values[1]}).Add(count)
	prom.BytesUploadedToS3.With(prometheus.Labels{"appId": values[0], "integration": values[1]}).Add(size)
	// Successfully uploaded, clear the buffer
	c.queue[key] = make([][]byte, 0)
	return nil
}

// Close will close down the DeadLetter Queue, flushing its content to S3.
func (c *Client) Close() {
	logger := log.WithField("function", "Close")
	logger.Info("Closing Deadletter Queue")
	// Wait for any last add funcs to finish
	c.addingWaitGroup.Wait()
	if err := c.Flush(); err != nil {
		logger.Error(err)
	}
	logger.Info("Closed Deadletter Queue")
}

func (c *Client) Length(key string) int {
	return len(c.queue[key])
}
