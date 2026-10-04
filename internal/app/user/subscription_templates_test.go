package user

import (
	"github.com/flosch/pongo2/v6"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readBundledTemplateForTest loads a template file shipped in the panel's
// templates directory, walking upwards from the test working directory.
func readBundledTemplateForTest(t *testing.T, name string) (string, error) {
	t.Helper()
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 4; i++ {
			candidates = append(candidates, filepath.Join(dir, "templates", filepath.FromSlash(name)))
			dir = filepath.Dir(dir)
		}
	}
	for _, candidate := range candidates {
		if data, err := os.ReadFile(candidate); err == nil {
			return string(data), nil
		}
	}
	return "", os.ErrNotExist
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

func containsAny(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

// retiredIdentityNeedles returns the branding the rendered subscription page
// must never contain: the fork it was ported from, and the replaced mark with
// its cut-out asset.
//
// The author line is required here, not retired, so it is asserted as a
// marker below rather than listed as a needle.
//
// The fragments are concatenated at run time so this file never spells the
// retired strings out in full; Web/tools/check-brand-geometry.cjs scans every
// source file for them, and a literal here would make the guard flag its own
// defensive assertion.
func retiredIdentityNeedles() []string {
	return []string{
		"BaToHub",
		"bato_theme",
		"squared-off " + "G",
		"gamaj" + "-cut",
	}
}

// The Gamaj sub templates ship with the panel and must always render with the
// standard subscription context (username, links, expire, support URL).
func TestGamajSubscriptionTemplatesRender(t *testing.T) {
	for _, name := range []string{"subscription/gamaj-green.html", "subscription/gamaj-sub.html"} {
		content, err := readBundledTemplateForTest(t, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		tpl, err := pongo2.FromString(normalizeLegacySubscriptionTemplate(content))
		if err != nil {
			t.Fatalf("%s: template parse: %v", name, err)
		}
		user := UserDetail{Username: "buyer", Status: "active", SubscriptionURL: "https://panel.example/sub/buyer"}
		user.Expire = new(int64)
		*user.Expire = 1800000000
		out, err := tpl.Execute(subscriptionTemplateContext(
			user,
			[]string{"vless://example", "vmess://example"},
			"https://panel.example/sub/buyer/usage",
			"https://t.me/support",
			"token",
		))
		if err != nil {
			t.Fatalf("%s: render: %v", name, err)
		}
		for _, marker := range []string{"buyer", "vless://example", "Coded by AsliCode", "GAMAJ"} {
			if !contains(out, marker) {
				t.Fatalf("%s: rendered output missing %q", name, marker)
			}
		}
		if containsAny(out, retiredIdentityNeedles()...) {
			t.Fatalf("%s: retired identity leaked into rendered page", name)
		}
	}
}
