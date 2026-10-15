package migrations

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedMigrationsAreAvailable(t *testing.T) {
	entries, err := fs.ReadDir(files, "migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one embedded SQL migration")
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			t.Errorf("unexpected embedded migration entry %q", entry.Name())
		}
	}
}
