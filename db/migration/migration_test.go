package migration

import (
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

// This tests is coming from this PR
// https://github.com/tinkerbell/tink/pull/466
//
// We didn't fully understood how
// the migration library sorts migration before having to troubleshoot
// that PR. Even if we declare them in order as part of the GetMigrations
// function the library still applies quick sort on the migration ID to
// stay concurrency safe
//
// https://github.com/rubenv/sql-migrate/blob/011dc47c6043b25483490739b61cabbc5da7ee9a/migrate.go#L216
//
// In order to improve the visibility and to acknowledge that the way we
// append migrations to GetMigrations is the same we get after the sorting
// we wrote this tests.
// A possible solution is to write our own MigrationSet who won't do the
// ordering. In the meantime I (gianarb) will start a conversation with library
// maintainer to figure out how we can improve this.
func TestMigrationsAreOrderendInTheWayWeDeclaredThem(t *testing.T) {
	declaredMigration := GetMigrations().Migrations
	orderedMigration, err := GetMigrations().FindMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(declaredMigration) != len(orderedMigration) {
		t.Error("Migrations do not have the same number of elements. This should never happen.")
	}

	for ii, dm := range declaredMigration {
		if dm.Id != orderedMigration[ii].Id {
			t.Errorf("Expected migration \"%s\" but got \"%s\"", dm.Id, orderedMigration[ii].Id)
		}
	}
}

func TestMigrationFuncNamesMatchIDs(t *testing.T) {
	timestamps := map[string]bool{}
	for _, m := range migrations {
		pc := reflect.ValueOf(m).Pointer()
		fn := runtime.FuncForPC(pc)

		file, _ := fn.FileLine(pc)
		file = path.Base(file)
		root := strings.TrimSuffix(file, path.Ext(file))

		migration := m()
		mid := strings.Split(migration.Id, "-")[0]

		assert.Equal(t, root, migration.Id, "file root name (%s) and migration id (%s) do not match", root, migration.Id)

		fnName := strings.Split(fn.Name(), "migration.")[1]
		assert.Equal(t, fnName, "Get"+mid, "migration func name and id timestamp mismatch, func: %s, migration.Id: %s", fnName, mid)

		assert.NotContains(t, timestamps, mid)
	}
}

func TestGetMigrations(t *testing.T) {
	m := GetMigrations()
	assert.Len(t, m.Migrations, len(migrations))
}
