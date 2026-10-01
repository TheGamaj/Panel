package api

import "testing"

func TestLoadConfigReadsGamajDatabaseURL(t *testing.T) {
	t.Setenv("GAMAJ_DATABASE_URL", "sqlite:///gamaj.db")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database != "sqlite:///gamaj.db" {
		t.Fatalf("Database=%q want %q", cfg.Database, "sqlite:///gamaj.db")
	}
}

func TestLoadConfigRequiresGamajDatabaseURL(t *testing.T) {
	t.Setenv("GAMAJ_DATABASE_URL", "")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() succeeded without GAMAJ_DATABASE_URL")
	}
}

func TestLoadConfigUsesRecordingDefaultsFromDatabaseSettings(t *testing.T) {
	t.Setenv("GAMAJ_DATABASE_URL", "sqlite:///usage-flags.db")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RecordNodeUsage {
		t.Fatal("RecordNodeUsage=false want true")
	}
	if !cfg.RecordNodeUserUsages {
		t.Fatal("RecordNodeUserUsages=false want true")
	}
}
