package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tinkerbell/tink/db"
)

type NewPostgresDatabaseRequest struct {
	ApplyMigration bool
}

// NewPostgresDatabaseClient returns a SQL client ready to be used. Behind the
// scene it is starting a Docker container that will get cleaned up when the
// test is over. Tests using this function are safe to run in parallel
func NewPostgresDatabaseClient(t *testing.T, ctx context.Context, req NewPostgresDatabaseRequest) (*sql.DB, *db.TinkDB, func() error) {
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:13.1",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
			Env: map[string]string{
				"POSTGRES_PASSWORD": "tinkerbell",
				"POSTGRES_USER":     "tinkerbell",
				"POSTGRES_DB":       "tinkerbell",
			},
		},
		Started: true,
	})
	if err != nil {
		t.Error(err)
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		t.Error(err)
	}
	dbCon, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=localhost port=%d user=%s password=%s dbname=%s sslmode=disable",
			port.Int(),
			"tinkerbell",
			"tinkerbell",
			"tinkerbell"))
	if err != nil {
		t.Error(err)
	}

CHECK_DB:
	for ii := 0; ii < 5; ii++ {
		err = dbCon.Ping()
		if err != nil {
			t.Log(errors.Wrap(err, "db check"))
			time.Sleep(1 * time.Second)
			goto CHECK_DB
		}
	}
	if err != nil {
		t.Fatal(err)
	}

	tinkDB := db.Connect(dbCon, log.Test(t, "db-test"))
	if req.ApplyMigration {
		n, err := tinkDB.Migrate()
		if err != nil {
			t.Error(err)
		}
		t.Log(fmt.Sprintf("applied %d migrations", n))
	}
	return dbCon, tinkDB, func() error {
		return postgresC.Terminate(ctx)
	}
}
