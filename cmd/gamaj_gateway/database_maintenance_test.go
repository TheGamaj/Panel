package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunManagedDatabaseMaintenance(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is not available on Windows")
	}

	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := filepath.Join(dir, "gamaj")
	contents := "#!/bin/sh\nprintf '%s' \"$*\" > \"" + argsFile + "\"\n"
	if err := os.WriteFile(script, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}

	previousPath := gamajScriptPath
	gamajScriptPath = script
	t.Cleanup(func() { gamajScriptPath = previousPath })

	runManagedDatabaseMaintenance(context.Background())
	got, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != "database-maintenance" {
		t.Fatalf("unexpected arguments: %q", got)
	}
}

func TestRunManagedDatabaseMaintenanceSkipsMissingInstaller(t *testing.T) {
	previousPath := gamajScriptPath
	gamajScriptPath = filepath.Join(t.TempDir(), "missing")
	t.Cleanup(func() { gamajScriptPath = previousPath })

	runManagedDatabaseMaintenance(context.Background())
}
