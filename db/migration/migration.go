package migration

import (
	migrate "github.com/rubenv/sql-migrate"
)

var migrations = []func() *migrate.Migration{
	Get202009171251,
	Get202010071530,
	Get202010221010,
	Get202012041103,
	Get202012091055,
	Get202012169135,
}

func GetMigrations() *migrate.MemoryMigrationSource {
	m := make([]*migrate.Migration, len(migrations))
	for i := range migrations {
		m[i] = migrations[i]()
	}
	return &migrate.MemoryMigrationSource{
		Migrations: m,
	}
}
