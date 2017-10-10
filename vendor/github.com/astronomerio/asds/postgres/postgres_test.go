package postgres

import (
	"testing"
	"github.com/sirupsen/logrus"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"fmt"
)

var (
	dbConfig = &ConnectionInfo{
		ProvisionerUser: "postgres",
		ProvisionerPw: "password",
		Hostname: "localhost",
		Port: "5432",
		DbName: "astronomer_saas",
	}
	log = logrus.New()
	c = &Client{
		log: log,
		info: dbConfig,
	}

)

func TestGetConnectionString(t *testing.T) {
	got := c.GetConnectString()
	expected := "user=postgres password=password dbname=astronomer_saas host=localhost sslmode=disable"
	if got != expected {
		t.Errorf("was expecting %s got %s", expected, got)
	}
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s",err.Error())
	}
	defer db.Close()

	// Wild card match random password
	query := fmt.Sprintf(CREATE_USER, "123", "(.+)", "123")
	mock.ExpectQuery(query).WillReturnRows(&sqlmock.Rows{})

	user, _, err := c.createUser(db, "123")
	if err != nil {
		t.Fatalf("error creating user %s", err.Error())
	}
	if user != "user_123" {
		t.Fatalf("error getting user name %s", err.Error())
	}
}

func TestCreateSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	query := fmt.Sprintf(CREATE_SCHEMA, "123", "123", "123", "123")
	mock.ExpectQuery(query).WillReturnRows(&sqlmock.Rows{})
	query = fmt.Sprintf(ALTER_SEARCH_PATH, "123", "123")
	mock.ExpectQuery(query).WillReturnRows(&sqlmock.Rows{})

	err = c.createSchema(db, "123")
	if err != nil {
		t.Fatalf("error creating schema %s", err.Error())
	}
}

func TestDropUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	mock.ExpectQuery("DROP USER user_123").WillReturnRows(&sqlmock.Rows{})

	err = c.dropUser(db, "123")
	if err != nil {
		t.Fatalf("error dropping use %s", err.Error())
	}
}

func TestDropSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	mock.ExpectQuery("DROP SCHEMA org_123 CASCADE").WillReturnRows(&sqlmock.Rows{})

	err = c.dropSchema(db, "123")
	if err != nil {
		t.Fatalf("error dropping schema %s", err.Error())
	}
}

func TestSchemaExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"schema_name"})
	rows.AddRow("org_123")
	mock.ExpectQuery(fmt.Sprintf(SCHEMA_EXISTS, "123")).WillReturnRows(rows)

	exists, err := c.schemaExists(db, "123")
	if err != nil {
		t.Fatalf("error checking if schema exists %s", err.Error())
	}
	if !exists {
		t.Errorf("org_%s should exist", "123")
	}

	rows = sqlmock.NewRows([]string{"schema_name"})
	mock.ExpectQuery(fmt.Sprintf(SCHEMA_EXISTS, "124")).WillReturnRows(rows)
	exists, err = c.schemaExists(db, "124")
	if err != nil {
		t.Fatalf("error checking if schema exists %s", err.Error())
	}
	if exists {
		t.Errorf("org_%s shouldn't exist", "124")
	}
}

func TestUserExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"rolname"})
	rows.AddRow("user_123")
	mock.ExpectQuery(fmt.Sprintf(GET_ROLES, "123")).WillReturnRows(rows)
	exists, err := c.userExists(db, "123")
	if err != nil {
		t.Fatalf("error checking is user exists %s", err.Error())
	}
	if !exists {
		t.Errorf("user %s should exist", "123")
	}

	rows = sqlmock.NewRows([]string{"rolname"})
	mock.ExpectQuery(fmt.Sprintf(GET_ROLES, "124")).WillReturnRows(rows)
	exists, err = c.userExists(db, "124")
	if err != nil {
		t.Fatalf("error checking is user exists %s", err.Error())
	}
	if exists {
		t.Errorf("user %s shouldn't exist", "124")
	}
}

func TestGetAirflowPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	query := fmt.Sprintf(CREATE_USER, "123", "(.+)", "123")
	mock.ExpectQuery(query).WillReturnRows(&sqlmock.Rows{})
	query = fmt.Sprintf(CREATE_SCHEMA, "123", "123", "123", "123")
	mock.ExpectQuery(query).WillReturnRows(&sqlmock.Rows{})
	query = fmt.Sprintf(ALTER_SEARCH_PATH, "123", "123")
	mock.ExpectQuery(query).WillReturnRows(&sqlmock.Rows{})

	conn, err := c.GetAirflowPostgres(db, "123")
	if err != nil {
		t.Fatalf("error creating new postgres schema %s", err.Error())
	}
	if conn != "postgres://user_123:password@localhost:5432/astronomer_saas" {
		t.Errorf("%s is not formatted correctly", conn)
	}
}

func TestDestroyAirflowPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected failure creating mock databse %s", err.Error())
	}
	defer db.Close()

	// schema and user exist
	// schema exists
	rows := sqlmock.NewRows([]string{"schema_name"})
	rows.AddRow("org_123")
	mock.ExpectQuery(fmt.Sprintf(SCHEMA_EXISTS, "123")).WillReturnRows(rows)

	// drop schema
	mock.ExpectQuery("DROP SCHEMA org_123 CASCADE").WillReturnRows(&sqlmock.Rows{})

	// user exists
	rows = sqlmock.NewRows([]string{"rolname"})
	rows.AddRow("user_123")
	mock.ExpectQuery(fmt.Sprintf(GET_ROLES, "123")).WillReturnRows(rows)

	// drop user
	mock.ExpectQuery("DROP USER user_123").WillReturnRows(&sqlmock.Rows{})

	err = c.DestroyAirflowPostgres(db, "123")
	if err != nil {
		t.Fatalf("failed to destroy org_123 %s", err.Error())
	}

	// schema exists user doesn't
	rows = sqlmock.NewRows([]string{"schema_name"})
	rows.AddRow("org_123")
	mock.ExpectQuery(fmt.Sprintf(SCHEMA_EXISTS, "123")).WillReturnRows(rows)

	// drop schema
	mock.ExpectQuery("DROP SCHEMA org_123 CASCADE").WillReturnRows(&sqlmock.Rows{})

	// no user
	rows = sqlmock.NewRows([]string{"rolname"})
	mock.ExpectQuery(fmt.Sprintf(GET_ROLES, "123")).WillReturnRows(rows)

	err = c.DestroyAirflowPostgres(db, "123")
	if err != nil {
		t.Fatalf("failed to destroy org_123 %s", err.Error())
	}

	// schema doesn't exist user does
	rows = sqlmock.NewRows([]string{"schema_name"})
	mock.ExpectQuery(fmt.Sprintf(SCHEMA_EXISTS, "123")).WillReturnRows(rows)

	// user exists
	rows = sqlmock.NewRows([]string{"rolname"})
	rows.AddRow("user_123")
	mock.ExpectQuery(fmt.Sprintf(GET_ROLES, "123")).WillReturnRows(rows)

	// drop user
	mock.ExpectQuery("DROP USER user_123").WillReturnRows(&sqlmock.Rows{})

	err = c.DestroyAirflowPostgres(db, "123")
	if err != nil {
		t.Fatalf("failed to destroy org_123 %s", err.Error())
	}
}