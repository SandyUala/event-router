package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("package", "cassandra")
)

type Client struct {
	session *gocql.Session
	configs *Configs
}

type Configs struct {
	Servers           []string
	Keyspace          string
	ReplicationFactor int
	MessageTableName  string
	Consistency       gocql.Consistency
}

func NewCilent(configs *Configs) (*Client, error) {
	cluster := gocql.NewCluster(configs.Servers...)
	cluster.Keyspace = configs.Keyspace
	cluster.Consistency = configs.Consistency
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating cassandra cluster session")
	}
	client := &Client{
		session: session,
		configs: configs,
	}
	if err := client.initMessagesTable(configs.MessageTableName); err != nil {
		return nil, errors.Wrap(err, "Error initializing message table")
	}
	return client, nil
}

func (c *Client) initMessagesTable(tableName string) error {
	logger := log.WithField("function", "initMessageTables")
	logger.Debug("Entered initMessageTables")

	// Create Keyspace
	keyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s"+
			"WITH replication = {'class':'SimpleStrategy',"+
			"'replication_factor': %d", c.configs.Keyspace, c.configs.ReplicationFactor)
	if err := c.session.Query(keyspaceQuery).Exec(); err != nil {
		return errors.Wrap(err, "error attempting to create cassandra keyspace")
	}
	logger.Debug("Created Keyspace")

	/*
		Cassandra Table for Messages

		| Run ID | Message ID |  Task ID |
		|--------|------------|----------|
		|   INT  |    TEXT    |   TEXT   |

		PRIMARY KEY (runID, taskID, messageID)
	*/
	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
	runID int,
	taskID text,
	messageID text,
	PRIMARY KEY (runID, taskID, messageID)`, c.configs.Keyspace, c.configs.MessageTableName)
	if err := c.session.Query(createTableQuery).Exec(); err != nil {
		return errors.Wrap(err, "error creating messages table")
	}
	logger.Debug("Created messages table")

	return nil
}
