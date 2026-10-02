package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultSponsorManifestURL = "https://raw.githubusercontent.com/TheGamaj/Panel/dev/sponsors/gamaj/manifest.json"

type Config struct {
	Database                     string
	CertificateBase              string
	CertbotBinary                string
	ExternalAppsBase             string
	SponsorManifestURL           string
	SponsorCacheDir              string
	MySQLRootPassword            string
	NodeOperationsPollInterval   string
	NodeUsageCollectionInterval  string
	NodeUsageCollectionLimit     int
	NodeUsageFlushInterval       string
	NodeUsageFlushBatchSize      int
	RecordNodeUsage              bool
	RecordNodeUserUsages         bool
	AdminLifecycleInterval       string
	UserLifecycleInterval        string
	UserLifecycleBatchSize       int
	UserUsageResetInterval       string
	UserUsageResetBatchSize      int
	UserAutodeleteInterval       string
	UserAutodeleteBatchSize      int
	UsersAutodeleteDays          int
	UserAutodeleteIncludeLimited bool
	JWTAccessTokenExpireMinutes  int
	UsersListTimeoutSeconds      float64
	SubscriptionReadOnly         bool
	TelegramAPIBase              string
	APIDocsEnabled               bool
	WebhookAddresses             []string
	WebhookSecret                string
	WebhookSendInterval          string
	WebhookMaxRetries            int
	WebhookRetryInterval         string
}

func LoadConfig() (Config, error) {
	env := loadEnvFiles()
	lookup := func(keys ...string) string {
		for _, key := range keys {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				return value
			}
			if value := strings.TrimSpace(env[key]); value != "" {
				return value
			}
		}
		return ""
	}

	cfg := Config{
		Database:                     lookup("GAMAJ_DATABASE_URL"),
		CertificateBase:              lookup("GAMAJ_CERT_BASE"),
		CertbotBinary:                lookup("GAMAJ_CERTBOT_BIN"),
		ExternalAppsBase:             lookup("GAMAJ_EXTERNAL_APPS_BASE"),
		SponsorManifestURL:           firstNonEmpty(lookup("GAMAJ_SPONSOR_MANIFEST_URL"), defaultSponsorManifestURL),
		SponsorCacheDir:              firstNonEmpty(lookup("GAMAJ_SPONSOR_CACHE_DIR"), filepath.Join(firstNonEmpty(lookup("GAMAJ_DATA_DIR"), "/var/lib/gamaj"), "sponsor-cache")),
		MySQLRootPassword:            lookup("GAMAJ_MYSQL_ROOT_PASSWORD"),
		NodeOperationsPollInterval:   lookup("GAMAJ_NODE_OPERATIONS_POLL_INTERVAL"),
		NodeUsageCollectionInterval:  lookup("GAMAJ_NODE_USAGE_COLLECTION_INTERVAL"),
		NodeUsageCollectionLimit:     parseIntDefault(lookup("GAMAJ_NODE_USAGE_COLLECTION_LIMIT"), 0),
		NodeUsageFlushInterval:       lookup("GAMAJ_NODE_USAGE_FLUSH_INTERVAL"),
		NodeUsageFlushBatchSize:      parseIntDefault(lookup("GAMAJ_NODE_USAGE_FLUSH_BATCH_SIZE"), 2000),
		RecordNodeUsage:              true,
		RecordNodeUserUsages:         true,
		AdminLifecycleInterval:       lookup("GAMAJ_ADMIN_LIFECYCLE_INTERVAL"),
		UserLifecycleInterval:        lookup("GAMAJ_USER_LIFECYCLE_INTERVAL"),
		UserLifecycleBatchSize:       parseIntDefault(lookup("GAMAJ_USER_LIFECYCLE_BATCH_SIZE"), 500),
		UserUsageResetInterval:       lookup("GAMAJ_USER_USAGE_RESET_INTERVAL"),
		UserUsageResetBatchSize:      parseIntDefault(lookup("GAMAJ_USER_USAGE_RESET_BATCH_SIZE"), 500),
		UserAutodeleteInterval:       lookup("GAMAJ_USER_AUTODELETE_INTERVAL"),
		UserAutodeleteBatchSize:      parseIntDefault(lookup("GAMAJ_USER_AUTODELETE_BATCH_SIZE"), 500),
		UsersAutodeleteDays:          parseIntDefault(lookup("GAMAJ_USER_AUTODELETE_DAYS"), -1),
		UserAutodeleteIncludeLimited: parseBoolDefault(lookup("GAMAJ_USER_AUTODELETE_INCLUDE_LIMITED"), false),
		JWTAccessTokenExpireMinutes:  parseIntDefault(lookup("GAMAJ_JWT_ACCESS_TOKEN_EXPIRE_MINUTES"), 1440),
		UsersListTimeoutSeconds:      parseFloatDefault(lookup("GAMAJ_USERS_LIST_TIMEOUT_SECONDS"), 0),
		TelegramAPIBase:              lookup("GAMAJ_TELEGRAM_API_BASE"),
		WebhookAddresses:             splitWebhookAddresses(lookup("GAMAJ_WEBHOOK_ADDRESS")),
		WebhookSecret:                lookup("GAMAJ_WEBHOOK_SECRET"),
		WebhookSendInterval:          lookup("GAMAJ_WEBHOOK_SEND_INTERVAL"),
		WebhookMaxRetries:            parseIntDefault(lookup("GAMAJ_WEBHOOK_MAX_RETRIES"), 3),
		WebhookRetryInterval:         lookup("GAMAJ_WEBHOOK_RETRY_INTERVAL"),
	}
	if cfg.Database == "" {
		return Config{}, fmt.Errorf("GAMAJ_DATABASE_URL is required")
	}
	return cfg, nil
}

// splitWebhookAddresses parses GAMAJ_WEBHOOK_ADDRESS, which may list several endpoints
// separated by commas or whitespace.
func splitWebhookAddresses(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseBoolDefault(value string, fallback bool) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parseIntDefault(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return fallback
	}
	return result
}

func parseFloatDefault(value string, fallback float64) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return result
}

func loadEnvFiles() map[string]string {
	result := map[string]string{}
	for _, path := range candidateEnvFiles() {
		mergeEnvFile(result, path)
	}
	return result
}

func candidateEnvFiles() []string {
	seen := map[string]bool{}
	add := func(paths []string, path string) []string {
		path = strings.TrimSpace(path)
		if path == "" {
			return paths
		}
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
		if seen[path] {
			return paths
		}
		seen[path] = true
		return append(paths, path)
	}

	paths := []string{}
	paths = add(paths, os.Getenv("GAMAJ_ENV_FILE"))
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		paths = add(paths, filepath.Join(dir, ".env"))
		paths = add(paths, filepath.Join(filepath.Dir(dir), ".env"))
	}
	if cwd, err := os.Getwd(); err == nil {
		paths = add(paths, filepath.Join(cwd, ".env"))
		paths = add(paths, filepath.Join(filepath.Dir(cwd), ".env"))
	}
	return paths
}

func mergeEnvFile(dst map[string]string, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key != "" {
			if _, exists := dst[key]; exists {
				continue
			}
			dst[key] = value
		}
	}
}
