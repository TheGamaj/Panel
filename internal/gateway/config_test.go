package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigUsesGamajEnv(t *testing.T) {
	t.Setenv("GAMAJ_GATEWAY_ADDR", "")
	t.Setenv("GAMAJ_HOST", "127.0.0.1")
	t.Setenv("GAMAJ_PORT", "9443")
	t.Setenv("GAMAJ_SSL_CERTFILE", "/tmp/gamaj/fullchain.pem")
	t.Setenv("GAMAJ_SSL_KEYFILE", "/tmp/gamaj/key.pem")

	cfg := LoadConfig()

	if cfg.Addr != "127.0.0.1:9443" {
		t.Fatalf("Addr=%q want %q", cfg.Addr, "127.0.0.1:9443")
	}
	if cfg.TLSCertFile != "/tmp/gamaj/fullchain.pem" {
		t.Fatalf("TLSCertFile=%q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "/tmp/gamaj/key.pem" {
		t.Fatalf("TLSKeyFile=%q", cfg.TLSKeyFile)
	}
}

func TestLoadConfigDefaultsToPort616(t *testing.T) {
	t.Setenv("GAMAJ_GATEWAY_ADDR", "")
	t.Setenv("GAMAJ_HOST", "")
	t.Setenv("GAMAJ_PORT", "")

	cfg := LoadConfig()

	if cfg.Addr != ":616" {
		t.Fatalf("Addr=%q want %q", cfg.Addr, ":616")
	}
}

func TestLoadConfigKeepsGatewayAddrOverride(t *testing.T) {
	t.Setenv("GAMAJ_GATEWAY_ADDR", ":18080")
	t.Setenv("GAMAJ_HOST", "127.0.0.1")
	t.Setenv("GAMAJ_PORT", "9443")

	cfg := LoadConfig()

	if cfg.Addr != ":18080" {
		t.Fatalf("Addr=%q want %q", cfg.Addr, ":18080")
	}
}

func TestLoadConfigReadsGamajEnvFile(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	writeTestFile(t, envPath, `
GAMAJ_HOST = "127.0.0.1"
GAMAJ_PORT = "18083"
GAMAJ_SSL_CERTFILE = "/var/lib/gamaj/certs/fullchain.pem"
GAMAJ_SSL_KEYFILE = "/var/lib/gamaj/certs/key.pem"
`)
	t.Setenv("GAMAJ_ENV_FILE", envPath)
	t.Setenv("GAMAJ_GATEWAY_ADDR", "")
	t.Setenv("GAMAJ_HOST", "")
	t.Setenv("GAMAJ_PORT", "")
	t.Setenv("GAMAJ_SSL_CERTFILE", "")
	t.Setenv("GAMAJ_SSL_KEYFILE", "")

	cfg := LoadConfig()

	if cfg.Addr != "127.0.0.1:18083" {
		t.Fatalf("Addr=%q want %q", cfg.Addr, "127.0.0.1:18083")
	}
	if cfg.TLSCertFile != "/var/lib/gamaj/certs/fullchain.pem" {
		t.Fatalf("TLSCertFile=%q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "/var/lib/gamaj/certs/key.pem" {
		t.Fatalf("TLSKeyFile=%q", cfg.TLSKeyFile)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
