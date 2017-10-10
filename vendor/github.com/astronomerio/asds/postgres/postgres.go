package postgres

import (
	"crypto/rand"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	CREATE_USER   = "CREATE ROLE user_%s WITH LOGIN " +
		"NOSUPERUSER " +
		"NOCREATEDB " +
		"NOCREATEROLE " +
		"NOINHERIT " +
		"NOREPLICATION " +
		"CONNECTION " +
		"LIMIT -1 " +
		"PASSWORD '%s';" +
		"GRANT user_%s TO provisioner;"

	CREATE_SCHEMA = `CREATE SCHEMA org_%s AUTHORIZATION provisioner;
					 GRANT ALL ON SCHEMA org_%s TO user_%s;
					 ALTER DEFAULT PRIVILEGES FOR ROLE provisioner IN SCHEMA org_%s GRANT SELECT, DELETE ON TABLES TO provisioner;`
	ALTER_SEARCH_PATH = `ALTER ROLE user_%s SET search_path = org_%s`
	SCHEMA_EXISTS = `SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'org_%s';`
	GET_ROLES = `SELECT rolname FROM pg_roles WHERE rolname = 'user_%s';`
)

type PostgresClient interface {
	GetAirflowPostgres(db *sql.DB, orgID string) (string, error)
	DestroyAirflowPostgres(db *sql.DB, orgID string) error
	GetConnectString() string
}

type ConnectionInfo struct {
	ProvisionerUser string
	ProvisionerPw   string
	Hostname        string
	Port            string
	DbName          string
}

type Client struct {
	log  *logrus.Logger
	info *ConnectionInfo
}

func NewClient(info *ConnectionInfo, log *logrus.Logger) PostgresClient {
	return &Client{
		log:  log,
		info: info,
	}
}

func (c *Client) createUser(db *sql.DB, id string) (string, string, error) {
	logger := c.log.WithField("function", "createUser")
	logger.Debug("Creating user for organization: " + id)

	pw, err := generateRandomString(32)
	if err != nil {
		return "", "", err
	}
	query := fmt.Sprintf(CREATE_USER, id, pw, id)
	logger.WithField("query", query).Debug("About to execute create user query")
	_, err = db.Query(query)
	if err != nil {
		return "", "", err
	}

	return "user_" + id, pw, nil
}

func (c *Client) createSchema(db *sql.DB, id string) error {
	logger := c.log.WithField("function", "createSchema")
	logger.Debug("Creating schema for organization: " + id)
	_, err := db.Query(fmt.Sprintf(CREATE_SCHEMA, id, id, id, id))
	if err != nil {
		return err
	}
	_, err = db.Query(fmt.Sprintf(ALTER_SEARCH_PATH, id, id))
	if err != nil {
		logger.Debug("Creating of schema: " + id + "has failed")
		logger.Debug(err.Error())
		return err
	}

	return nil
}

func (c *Client) dropUser(db *sql.DB, id string) error {
	logger := c.log.WithField("function", "dropUser")
	logger.Debug("Dropping user for for organization: " + id)

	_, err := db.Query("DROP USER " + "user_" + id)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) dropSchema(db *sql.DB, id string) error {
	logger := c.log.WithField("function", "dropSchema")
	logger.Debug("Dropping schema for for organization: " + id)

	_, err := db.Query("DROP SCHEMA " + "org_" + id + " CASCADE")
	if err != nil {
		return err
	}

	return nil
}

// GetAirflowPostgres creates a new user and schema for an organization and returns a connection string.
// The connection string is formatted as such postgres://userName:pw@hostName5432/astronomer_saas
func (c *Client) GetAirflowPostgres(db *sql.DB, orgID string) (string, error) {
	logger := c.log.WithField("function", "GetAirflowPostgres")
	logger.Debug("Entered GetAirflowPostgres")

	user, pw, err := c.createUser(db, orgID)
	if err != nil {
		return "", err
	}

	err = c.createSchema(db, orgID)
	if err != nil {
		return "", err
	}
	connectionString := "postgres://" + user + ":" + pw + "@" + c.info.Hostname + ":" + c.info.Port + "/" + c.info.DbName

	return connectionString, nil
}

// DestroyAirflowPostgres will remove the organization's schema and use from the astronoemr_saas database. If one or
// the other doesn't exist because of a partial provision or manual deletion it'll still clean up the other.
func (c *Client) DestroyAirflowPostgres(db *sql.DB, orgID string) error {
	logger := c.log.WithField("function", "DestroyAirflowPostgres")
	logger.Debug("Entered DestroyAirflowPostgres")

	exists, err := c.schemaExists(db, orgID)
	if err != nil {
		return err
	}
	if exists {
		err = c.dropSchema(db, orgID)
		if err != nil {
			return err
		}
	} else {
		logger.Debug("org_" + orgID + " doesn't exist")
	}

	exists, err = c.userExists(db, orgID)
	if err != nil {
		return err
	}
	if exists {
		err = c.dropUser(db, orgID)
		if err != nil {
			return err
		}
	} else {
		logger.Debug("user_" + orgID + " doesn't exist")
	}

	return nil
}

func (c *Client) schemaExists(db *sql.DB, orgID string) (bool, error) {
	logger := c.log.WithField("function", "schemaExists")
	logger.Debug("Entered schemaExists")

	rows, err := db.Query(fmt.Sprintf(SCHEMA_EXISTS, orgID))
	if err != nil {
		return false, err
	}
	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return false, err
		}
	}
	if name == "org_" + orgID {
		return true, nil
	}

	return false, nil
}

func (c *Client) userExists(db *sql.DB, orgID string) (bool, error) {
	logger := c.log.WithField("function", "userExists")
	logger.Debug("Entered userExists")
	rows, err := db.Query(fmt.Sprintf(GET_ROLES, orgID))
	defer rows.Close()
	if err != nil {
		return false, err
	}
	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return false, err
		}
	}
	if name == "user_" + orgID {
		return true, nil
	}
	return false, nil
}

func (c *Client) GetConnectString() string {
	logger := c.log.WithField("GetConnectString", "DestroyAirflowPostgres")
	logger.Debug("GetConnectString")
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable",
		c.info.ProvisionerUser, c.info.ProvisionerPw, c.info.DbName, c.info.Hostname)
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
