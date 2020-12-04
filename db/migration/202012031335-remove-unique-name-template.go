package migration

import migrate "github.com/rubenv/sql-migrate"

func Get202012031335() *migrate.Migration {
	return &migrate.Migration{
		Id: "202012031335-remove-unique-name-template",
		Up: []string{`
		ALTER TABLE template DROP CONSTRAINT template_name_key;
`},
	}
}
