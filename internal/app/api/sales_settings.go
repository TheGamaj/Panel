package api

import (
	"fmt"
	"net/http"
	"strings"

	settingsapp "github.com/TheGamaj/Panel/internal/app/settings"
)

func fmtSscan(text string, target any) (int, error) {
	return fmt.Sscan(text, target)
}

// handleSalesSettings manages the Tetraminator gateway configuration
// (GET returns a masked view; PUT updates it). Only sudo admins reach this
// handler through the settings route group.
func (s *Server) handleSalesSettings(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/settings/sales" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		masked, err := s.settingsRepo.MaskedSalesSettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, masked)
	case http.MethodPut:
		raw, err := decodeRawJSONMap(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		update := settingsapp.SalesSettingsUpdate{}
		if value, ok := raw["tetraminator_enabled"]; ok {
			enabled := strings.EqualFold(strings.Trim(string(value), `"`), "true") || strings.Contains(string(value), "true")
			update.TetraminatorEnabled = &enabled
		}
		if value, ok := raw["tetraminator_base_url"]; ok {
			text := strings.Trim(string(value), `"`)
			update.TetraminatorBaseURL = &text
		}
		if value, ok := raw["tetraminator_api_key"]; ok {
			text := strings.Trim(string(value), `"`)
			update.TetraminatorAPIKey = &text
		}
		if value, ok := raw["tetraminator_callback_url"]; ok {
			text := strings.Trim(string(value), `"`)
			update.TetraminatorCallback = &text
		}
		if value, ok := raw["tetraminator_min_price"]; ok {
			var parsed int64
			_, _ = fmtSscan(strings.Trim(string(value), `"`), &parsed)
			update.TetraminatorMinPrice = &parsed
		}
		if value, ok := raw["tetraminator_max_price"]; ok {
			var parsed int64
			_, _ = fmtSscan(strings.Trim(string(value), `"`), &parsed)
			update.TetraminatorMaxPrice = &parsed
		}
		masked, err := s.settingsRepo.UpdateSalesSettings(r.Context(), update)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, masked)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
