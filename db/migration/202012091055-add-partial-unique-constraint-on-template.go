package migration

import migrate "github.com/rubenv/sql-migrate"

func Get202012091055() *migrate.Migration {
	return &migrate.Migration{
		Id: "202012091055-add-partial-unique-constraint-on-template",
		Up: []string{`
                CREATE UNIQUE INDEX IF NOT EXISTS uidx_template_name ON template (name) WHERE deleted_at IS NULL;
                `},
	}
}
