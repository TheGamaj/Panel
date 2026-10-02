package settings

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// SalesSettings reads the Gamaj sales configuration (Tetraminator gateway).
// The API key is returned only to internal callers that need to create
// invoices; the dashboard API never exposes it.
func (r Repository) SalesSettings(ctx context.Context) (SalesSettings, error) {
	if err := r.ensureRuntimeSettingsRecord(ctx); err != nil {
		return SalesSettings{}, err
	}
	var result SalesSettings
	err := r.db.QueryRowContext(ctx, `
SELECT
	COALESCE(tetraminator_enabled, 0),
	COALESCE(tetraminator_base_url, ''),
	COALESCE(tetraminator_api_key, ''),
	COALESCE(tetraminator_min_price, 50000),
	COALESCE(tetraminator_max_price, 10000000),
	COALESCE(tetraminator_callback_path, '')
FROM settings
WHERE id = 1
LIMIT 1`).Scan(
		&result.TetraminatorEnabled,
		&result.TetraminatorBaseURL,
		&result.TetraminatorAPIKey,
		&result.TetraminatorMinPrice,
		&result.TetraminatorMaxPrice,
		&result.TetraminatorCallback,
	)
	if err != nil {
		return SalesSettings{}, err
	}
	result.TetraminatorBaseURL = strings.TrimRight(strings.TrimSpace(result.TetraminatorBaseURL), "/")
	return result, nil
}

// MaskedSalesSettings returns the sales configuration with the gateway API
// key reduced to its last four characters.
func (r Repository) MaskedSalesSettings(ctx context.Context) (MaskedSalesSettings, error) {
	sales, err := r.SalesSettings(ctx)
	if err != nil {
		return MaskedSalesSettings{}, err
	}
	masked := ""
	if sales.TetraminatorAPIKey != "" {
		if len(sales.TetraminatorAPIKey) <= 4 {
			masked = "****"
		} else {
			masked = "****" + sales.TetraminatorAPIKey[len(sales.TetraminatorAPIKey)-4:]
		}
	}
	return MaskedSalesSettings{
		TetraminatorEnabled:     sales.TetraminatorEnabled,
		TetraminatorConfigured:  sales.TetraminatorBaseURL != "" && sales.TetraminatorAPIKey != "",
		TetraminatorMaskedKey:   masked,
		TetraminatorBaseURL:     sales.TetraminatorBaseURL,
		TetraminatorMinPrice:    sales.TetraminatorMinPrice,
		TetraminatorMaxPrice:    sales.TetraminatorMaxPrice,
		TetraminatorCallbackURL: sales.TetraminatorCallback,
	}, nil
}

// UpdateSalesSettings applies a partial Tetraminator settings update. An
// empty tetraminator_api_key keeps the stored key unchanged.
func (r Repository) UpdateSalesSettings(ctx context.Context, update SalesSettingsUpdate) (MaskedSalesSettings, error) {
	if err := r.ensureRuntimeSettingsRecord(ctx); err != nil {
		return MaskedSalesSettings{}, err
	}
	sets := []string{"updated_at = ?"}
	args := []any{dbTime(time.Now().UTC())}
	add := func(column string, value any) {
		sets = append(sets, column+" = ?")
		args = append(args, value)
	}
	if update.TetraminatorEnabled != nil {
		add("tetraminator_enabled", boolInt(*update.TetraminatorEnabled))
	}
	if update.TetraminatorBaseURL != nil {
		add("tetraminator_base_url", strings.TrimRight(strings.TrimSpace(*update.TetraminatorBaseURL), "/"))
	}
	if update.TetraminatorAPIKey != nil && strings.TrimSpace(*update.TetraminatorAPIKey) != "" {
		add("tetraminator_api_key", strings.TrimSpace(*update.TetraminatorAPIKey))
	}
	if update.TetraminatorMinPrice != nil && *update.TetraminatorMinPrice > 0 {
		add("tetraminator_min_price", *update.TetraminatorMinPrice)
	}
	if update.TetraminatorMaxPrice != nil && *update.TetraminatorMaxPrice > 0 {
		add("tetraminator_max_price", *update.TetraminatorMaxPrice)
	}
	if update.TetraminatorCallback != nil {
		add("tetraminator_callback_path", strings.TrimSpace(*update.TetraminatorCallback))
	}
	args = append(args, 1)
	if _, err := r.db.ExecContext(ctx, "UPDATE settings SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...); err != nil {
		return MaskedSalesSettings{}, err
	}
	return r.MaskedSalesSettings(ctx)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

var _ = sql.ErrNoRows
var _ = errors.New
