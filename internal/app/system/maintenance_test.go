package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Gamaj has one install path, so a legacy install-mode marker must never leak
// into the maintenance API.
func TestDefaultRuntimeDetectorAlwaysReportsBinaryInstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GAMAJ_INSTALL_MODE", "docker")
	t.Setenv("GAMAJ_APP_DIR", dir)
	t.Setenv("GAMAJ_BINARY_METADATA_FILE", filepath.Join(dir, "missing-release.json"))
	if err := os.WriteFile(filepath.Join(dir, ".install-mode"), []byte("docker\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	info := DefaultRuntimeDetector{}.Info()
	if info.Mode != "binary" || info.InstallMode != "binary" {
		t.Fatalf("mode=%q install_mode=%q, want binary", info.Mode, info.InstallMode)
	}
	if info.Image != BinaryRuntimeImage {
		t.Fatalf("image=%q, want %q", info.Image, BinaryRuntimeImage)
	}
}

func TestGitHubUpdateCheckerCachesSuccessfulStatus(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		switch r.URL.Path {
		case "/repos/TheGamaj/Panel/releases/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tag_name":     "v0.2.0",
				"name":         "v0.2.0",
				"published_at": "2026-06-24T00:00:00Z",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	now := time.Unix(1_780_000_000, 0)
	checker := &GitHubUpdateChecker{
		APIBase:    server.URL,
		HTTPClient: server.Client(),
		Now:        func() time.Time { return now },
		CacheTTL:   time.Hour,
		ErrorTTL:   time.Hour,
	}
	current := "v0.1.0"

	first := checker.Status(context.Background(), "TheGamaj/Panel", &current, "latest")
	second := checker.Status(context.Background(), "TheGamaj/Panel", &current, "latest")

	if first.Error != "" || second.Error != "" {
		t.Fatalf("unexpected errors: first=%q second=%q", first.Error, second.Error)
	}
	if first.Target == nil || *first.Target != "v0.2.0" {
		t.Fatalf("unexpected first target: %#v", first.Target)
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("expected one release request, got %d", got)
	}
}

func TestGitHubUpdateCheckerListsBuildsFromSwitchFloor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/TheGamaj/Panel/releases":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"tag_name": "v0.0.9"},
				{"tag_name": "v0.1.0", "published_at": "2026-06-24T00:00:00Z"},
				{"tag_name": "v0.1.0", "prerelease": true},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	checker := &GitHubUpdateChecker{
		APIBase:    server.URL,
		HTTPClient: server.Client(),
	}
	catalog, err := checker.Builds(context.Background(), "TheGamaj/Panel")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Floor != versionSwitchFloor || len(catalog.Stable) != 1 || catalog.Stable[0].Version != "v0.1.0" {
		t.Fatalf("unexpected build catalog: %#v", catalog)
	}
}

func TestGitHubUpdateCheckerFallsBackToNewestPrerelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/TheGamaj/Panel/releases/latest":
			// GitHub hides prerelease publications from /releases/latest.
			http.NotFound(w, r)
		case "/repos/TheGamaj/Panel/releases":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"tag_name":     "is.0.0.1",
					"name":         "is.0.0.1",
					"prerelease":   true,
					"draft":        false,
					"published_at": "2026-09-29T09:57:51Z",
					"html_url":     "https://github.com/TheGamaj/Panel/releases/tag/is.0.0.1",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	checker := &GitHubUpdateChecker{
		APIBase:    server.URL,
		HTTPClient: server.Client(),
	}
	status := checker.Status(context.Background(), "TheGamaj/Panel", nil, "latest")

	if status.Error != "" {
		t.Fatalf("unexpected update error: %q", status.Error)
	}
	if status.LatestRelease == nil || stringFromAny((*status.LatestRelease)["tag"]) != "is.0.0.1" {
		t.Fatalf("expected prerelease fallback, got %#v", status.LatestRelease)
	}
}

func TestGitHubUpdateCheckerCachesErrors(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer server.Close()

	now := time.Unix(1_780_000_000, 0)
	checker := &GitHubUpdateChecker{
		APIBase:    server.URL,
		HTTPClient: server.Client(),
		Now:        func() time.Time { return now },
		CacheTTL:   time.Hour,
		ErrorTTL:   time.Hour,
	}

	first := checker.Status(context.Background(), "TheGamaj/Panel", nil, "latest")
	second := checker.Status(context.Background(), "TheGamaj/Panel", nil, "latest")

	if first.Error == "" || second.Error == "" {
		t.Fatalf("expected cached error, got first=%q second=%q", first.Error, second.Error)
	}
	// The failed /releases/latest probe and its /releases fallback both count
	// as one cached Status() failure.
	if got := atomic.LoadInt32(&requests); got != 2 {
		t.Fatalf("expected both failed requests to be cached, got %d", got)
	}
}
