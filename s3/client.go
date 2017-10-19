package s3

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("package", "s3")
)

type S3Client interface {
	SendToS3(bucket, key *string, body []byte) error
}

type Client struct {
	uploader *s3manager.Uploader
}

func NewClient() (*Client, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting AWS Session")
	}
	uploader := s3manager.NewUploader(sess)
	return &Client{
		uploader: uploader,
	}, nil
}

func (c *Client) SendToS3(bucket, key *string, body []byte) error {
	logger := log.WithField("function", "SendToS3")
	logger.WithFields(logrus.Fields{"bucket": bucket, "key": key}).Debug("Entered SendToS3")

	// Get a reader for the body
	r := bytes.NewReader(body)

	// We don't need the result, just if there is an error uploading
	_, err := c.uploader.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key:    key,
		Body:   r,
	})

	if err != nil {
		return errors.Wrap(err, "Error uploading to S3")
	}

	return nil
}

// Mock client for testing

type MockClient struct{}

func (c *MockClient) SendToS3(bucket, key *string, body []byte) error {
	return nil
}
