package api

import (
	"context"
	"database/sql"
	"time"

	adminapp "github.com/TheGamaj/Panel/internal/app/admin"
	externalapps "github.com/TheGamaj/Panel/internal/app/externalapps"
)

// botAPIKeyProvisioner returns the BotProvisioner used by the external apps
// manager: every Gamaj Bot installation receives a dedicated Gamaj API key
// minted for the first full-access admin so the bot can provision users
// exclusively through the Gamaj API. The plaintext key is returned once and
// stored only in the bot's own secret configuration file.
func (s *Server) botAPIKeyProvisioner() externalapps.BotProvisioner {
	return func(ctx context.Context, _ string) (string, error) {
		repository := adminapp.NewRepository(s.db, s.dialect)
		admin, found, err := repository.FirstFullAccessAdmin(ctx)
		if err != nil {
			return "", err
		}
		if !found {
			return "", adminapp.ErrAdminNotFound
		}
		var response apiKeyResponse
		err = s.withTx(ctx, func(tx *sql.Tx) error {
			var txErr error
			response, txErr = createAdminAPIKeyTx(ctx, tx, admin.ID, "", time.Now().UTC())
			return txErr
		})
		if err != nil {
			return "", err
		}
		if response.APIKey == nil || *response.APIKey == "" {
			return "", adminapp.ErrPermissionDenied
		}
		return *response.APIKey, nil
	}
}
