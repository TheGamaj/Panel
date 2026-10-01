package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	adminapp "github.com/TheGamaj/Panel/internal/app/admin"
)

// handleAdminWallet manages the reseller wallet of a single admin: GET
// returns the balance, PUT applies a ledgered balance delta. Quota limits
// (data + user count) are the standard admin limits already managed through
// the admins API.
func (s *Server) handleAdminWallet(w http.ResponseWriter, r *http.Request) {
	rest := walletAdminIDFromPath(r.URL.Path, "/api/admin/wallet/")
	if rest == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	var adminID int64
	for _, char := range rest {
		if char < '0' || char > '9' {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		adminID = adminID*10 + int64(char-'0')
	}
	if adminID <= 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	principal, ok := r.Context().Value(adminContextKey).(adminPrincipal)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing admin context")
		return
	}
	if principal.Context.Admin.Role != adminapp.RoleFullAccess && principal.Context.Admin.Role != adminapp.RoleSudo {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	switch r.Method {
	case http.MethodGet:
		var balance int64
		var settings string
		if err := s.db.QueryRowContext(r.Context(), `
SELECT COALESCE(reseller_wallet_balance, 0), COALESCE(reseller_settings, '{}') FROM admins WHERE id = ?`, adminID).Scan(&balance, &settings); err != nil {
			writeError(w, http.StatusNotFound, "admin not found")
			return
		}
		var settingsMap map[string]any
		_ = json.Unmarshal([]byte(settings), &settingsMap)
		if settingsMap == nil {
			settingsMap = map[string]any{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"balance": balance, "settings": settingsMap})
	case http.MethodPut:
		var payload struct {
			BalanceDelta int64             `json:"balance_delta"`
			Reason       string            `json:"reason"`
			Settings     map[string]string `json:"settings,omitempty"`
		}
		if err := decodeOptionalJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		err := s.withTx(r.Context(), func(tx *sql.Tx) error {
			var balance int64
			if err := tx.QueryRowContext(r.Context(), `SELECT COALESCE(reseller_wallet_balance, 0) FROM admins WHERE id = ?`, adminID).Scan(&balance); err != nil {
				return err
			}
			newBalance := balance + payload.BalanceDelta
			if newBalance < 0 {
				return statusError{status: http.StatusConflict, detail: "wallet balance would go negative"}
			}
			if _, err := tx.ExecContext(r.Context(), `UPDATE admins SET reseller_wallet_balance = ? WHERE id = ?`, newBalance, adminID); err != nil {
				return err
			}
			if payload.BalanceDelta != 0 {
				if _, err := tx.ExecContext(r.Context(), `
INSERT INTO reseller_wallet_transactions (admin_id, amount, balance_after, reason, reference, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
					adminID, payload.BalanceDelta, newBalance, walletReasonOrDefault(payload.Reason), "", dbTimestamp(time.Now().UTC())); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			writeStatusError(w, err)
			return
		}
		var balance int64
		_ = s.db.QueryRowContext(r.Context(), `SELECT COALESCE(reseller_wallet_balance, 0) FROM admins WHERE id = ?`, adminID).Scan(&balance)
		writeJSON(w, http.StatusOK, map[string]any{"balance": balance})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func walletReasonOrDefault(reason string) string {
	if reason == "" {
		return "admin_adjustment"
	}
	return reason
}

func walletAdminIDFromPath(path, prefix string) string {
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return ""
}

// handleAdminWalletRoot dispatches /api/admin/wallet/{id}.
func (s *Server) handleAdminWalletRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.handleAdminWallet(w, r)
}
