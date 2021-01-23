package migration

import (
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

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
