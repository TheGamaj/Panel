package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/TheGamaj/Panel/internal/app/logging"
)

var gamajScriptPath = "/usr/local/bin/gamaj"

// runManagedDatabaseMaintenance delegates managed database upkeep to the Gamaj
// installer script. Gamaj only ships the native binary install, so the presence
// of the installer is the single prerequisite.
func runManagedDatabaseMaintenance(ctx context.Context) {
	if _, err := os.Stat(gamajScriptPath); err != nil {
		return
	}

	maintenanceCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	output, err := exec.CommandContext(maintenanceCtx, gamajScriptPath, "database-maintenance").CombinedOutput()
	message := strings.TrimSpace(string(output))
	if err != nil {
		logging.Warnf(logging.ComponentRuntime, "managed database maintenance failed: %v output=%s", err, message)
		return
	}
	if message != "" {
		logging.Infof(logging.ComponentRuntime, "managed database maintenance: %s", message)
	}
}
