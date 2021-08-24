package migration

import migrate "github.com/rubenv/sql-migrate"

func Get202010221010() *migrate.Migration {
	return &migrate.Migration{
		Id: "202010221010-add-unique-index",
		Up: []string{`
CREATE UNIQUE INDEX IF NOT EXISTS uidx_workflow_worker_map ON workflow_worker_map (workflow_id, worker_id);
`},
	}
}
