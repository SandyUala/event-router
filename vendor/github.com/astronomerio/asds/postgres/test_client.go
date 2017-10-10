package postgres

import (
	"fmt"
	"database/sql"
)

type TestClient struct {
}

func NewTestClient() PostgresClient {
	return &TestClient{}
}

func (c *TestClient) GetAirflowPostgres(db *sql.DB, orgID string) (string, error) {
	return fmt.Sprintf("postgres://user_" + orgID + ":testpassword@testhost:5432/testdb"), nil
}

func (c *TestClient) DestroyAirflowPostgres(db *sql.DB, orgID string) error {
	return nil
}

func (c *TestClient) GetConnectString() string {
	return ""
}
