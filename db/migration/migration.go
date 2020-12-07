package migration

import migrate "github.com/rubenv/sql-migrate"

func GetMigrations() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			Get202009171251(),
			Get202010221010(),
			Get202012041103(),
		},
	}
}
