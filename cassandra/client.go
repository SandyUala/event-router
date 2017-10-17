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
	RunID             int64
}

func NewClient(configs *Configs) (*Client, error) {
	logger := log.WithField("function", "NewClient")
	logger.WithFields(logrus.Fields{"servers": configs.Servers, "keyspace": configs.Keyspace}).Debug("Entered New Client")

	// Create a session using the system key space so we can create our own key space
	systemClient := gocql.NewCluster(configs.Servers...)
	systemClient.Keyspace = "system"
	systemClient.Consistency = configs.Consistency
	systemSession, err := systemClient.CreateSession()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating cassandra cluster session")
	}
	if err = initKeyspace(systemSession, configs.Keyspace, configs.ReplicationFactor); err != nil {
		return nil, errors.Wrap(err, "Error initializing cluster session")
	}
	systemSession.Close()

	// Now create the needed session on the given key space
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
	if err := client.initMessagesTable(); err != nil {
		return nil, errors.Wrap(err, "Error initializing message table")
	}
	logger.Debug("Created Cassandra Client")
	return client, nil
}

func initKeyspace(session *gocql.Session, keyspace string, replicationFactor int) error {

	logger := log.WithField("function", "initMessageTables")
	logger.Debug("Entered initKeyspace")

	// Create Keyspace
	keyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s"+
			" WITH replication ="+
			" {'class':'SimpleStrategy','replication_factor': %d}", keyspace, replicationFactor)
	if err := session.Query(keyspaceQuery).Exec(); err != nil {
		return errors.Wrap(err, "error attempting to create cassandra keyspace")
	}
	logger.Debug("Created Keyspace")

	return nil
}

func (c *Client) initMessagesTable() error {
	logger := log.WithField("function", "initMessageTables")
	logger.Debug("Entered initMessageTables")

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
	messageID int,
	if err := c.session.Query(createTableQuery).Exec(); err != nil {
		return errors.Wrap(err, "error creating messages table")
	}
	logger.Debug("Created messages table")

	return nil
}

func (c *Client) InsertMessage(messageID string, taskID string) error {
	//logger := log.WithField("function", "InsertMessage")
	//logger.WithFields(logrus.Fields{"runID": c.configs.RunID, "messageID": messageID, "taskID": taskID}).Debug("Entered InsertMessage")

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s"+
			"(runID, taskID, messageID)"+
			"VALUES (?, ?, ?)", c.configs.MessageTableName)

	if err := c.session.Query(insertQuery, c.configs.RunID, taskID, messageID).Exec(); err != nil {
		return errors.Wrap(err, "Error inserting data")
	}
	return nil
}

func (c *Client) Close() {
	log.Info("Closing Cassandra")
	c.session.Close()
}
