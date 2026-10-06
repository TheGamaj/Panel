package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	adminapp "github.com/TheGamaj/Panel/internal/app/admin"
	salesapp "github.com/TheGamaj/Panel/internal/app/sales"
	userapp "github.com/TheGamaj/Panel/internal/app/user"
)

// botSalesPrincipal resolves the selling admin for every /bot request. The
// bot's Gamaj API key belongs to one admin; every sale is scoped to that
// admin so resellers only ever sell their own plans and see their own buyers.
func (s *Server) botSalesPrincipal(r *http.Request) (adminapp.Admin, error) {
	principal, ok := r.Context().Value(adminContextKey).(adminPrincipal)
	if !ok {
		return adminapp.Admin{}, errors.New("missing admin context")
	}
	admin := principal.Context.Admin
	if admin.ID <= 0 {
		return adminapp.Admin{}, errors.New("the Gamaj Bot API key has no admin owner")
	}
	return admin, nil
}

// writeBotError maps sales client errors onto HTTP responses.
func writeBotError(w http.ResponseWriter, err error) {
	var client salesapp.ClientError
	if errors.As(err, &client) {
		writeError(w, client.Status, client.Detail)
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}

// botSalesService lazily builds the sales service for this server with the
// configured Tetraminator gateway.
func (s *Server) botSalesService() *salesapp.Service {
	if s.salesService == nil {
		gateway := s.botTetraminator()
		s.salesService = salesapp.New(s.db, s.dialect, s.botProvisionUser, gateway)
	}
	return s.salesService
}

func (s *Server) botTetraminator() *salesapp.Tetraminator {
	sales, err := s.settingsRepo.SalesSettings(context.Background())
	if err != nil || !sales.TetraminatorEnabled {
		return nil
	}
	gateway := salesapp.NewTetraminator(salesapp.TetraminatorConfig{
		BaseURL: sales.TetraminatorBaseURL,
		APIKey:  sales.TetraminatorAPIKey,
	})
	if !gateway.Valid() {
		return nil
	}
	return gateway
}

// botProvisionUser creates the purchased user as the selling admin with the
// plan limits. It reuses the exact Go user core (permissions, quotas and
// provisioning) that the dashboard API uses.
func (s *Server) botProvisionUser(ctx context.Context, adminID int64, plan salesapp.Plan, buyerTelegramID, orderID string) (string, string, error) {
	repository := adminapp.NewRepository(s.db, s.dialect)
	admin, found, err := repository.AdminByID(ctx, adminID)
	if err != nil || !found {
		return "", "", fmt.Errorf("selling admin %d not found", adminID)
	}
	var dataLimit *int64
	if plan.DataLimit > 0 {
		dataLimit = &plan.DataLimit
	}
	var expire *int64
	if plan.DurationDays > 0 {
		value := time.Now().UTC().Add(time.Duration(plan.DurationDays) * 24 * time.Hour).Unix()
		expire = &value
	}
	var ipLimit *int64
	if plan.IPLimit > 0 {
		ipLimit = &plan.IPLimit
	}
	note := fmt.Sprintf("Gamaj Bot order %s", orderID)
	username := "bot-" + strings.ToLower(orderID[:12])
	payload := userapp.UserCreate{
		Username: username,
		UserPayloadBase: userapp.UserPayloadBase{
			DataLimit: dataLimit,
			Expire:    expire,
			IPLimit:   ipLimit,
			Note:      &note,
		},
	}
	if telegram := strings.TrimSpace(buyerTelegramID); telegram != "" {
		payload.TelegramID = &telegram
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}
	result, err := s.userService.CreateUser(ctx, admin, encoded)
	if err != nil {
		return "", "", err
	}
	detail, err := s.userService.UserGet(ctx, userapp.UserGetRequest{
		Username: result.Username,
		Admin: userapp.AdminContext{
			Username:       admin.Username,
			ID:             &admin.ID,
			Role:           string(admin.Role),
			CanViewTraffic: true,
			CanSortTraffic: true,
		},
	})
	if err != nil {
		return result.Username, "", nil
	}
	return result.Username, detail.SubscriptionURL, nil
}

type botPlanPayload struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	DurationDays int64  `json:"duration_days"`
	DataLimit    int64  `json:"data_limit"`
	IPLimit      int64  `json:"ip_limit"`
	ServiceID    int64  `json:"service_id"`
	Visible      *bool  `json:"visible"`
}

// handleBotPlans lists or creates the selling admin's sale plans.
func (s *Server) handleBotPlans(w http.ResponseWriter, r *http.Request) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:
		plans, err := s.botSalesService().Plans(r.Context(), admin.ID)
		if err != nil {
			writeBotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, plans)
	case http.MethodPost:
		var payload botPlanPayload
		if err := decodeBotJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(payload.Name) == "" || payload.Price <= 0 || payload.ServiceID <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "name, price and service_id are required")
			return
		}
		visible := true
		if payload.Visible != nil {
			visible = *payload.Visible
		}
		result, err := s.db.ExecContext(r.Context(), `
INSERT INTO sale_plans (admin_id, name, description, price, duration_days, data_limit_bytes, ip_limit, service_id, visible, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			admin.ID, strings.TrimSpace(payload.Name), strings.TrimSpace(payload.Description), payload.Price,
			defaultInt64(payload.DurationDays, 30), payload.DataLimit, payload.IPLimit, payload.ServiceID, visibleInt(visible),
			dbTimestamp(time.Now().UTC()), dbTimestamp(time.Now().UTC()))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		id, _ := result.LastInsertId()
		writeJSON(w, http.StatusCreated, map[string]any{"id": id})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func defaultInt64(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}

func visibleInt(visible bool) int {
	if visible {
		return 1
	}
	return 0
}

func decodeBotJSON(r *http.Request, target any) error {
	raw, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, 1<<20))
	if err != nil {
		return errors.New("invalid request body")
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return errors.New("invalid JSON body")
	}
	return nil
}

func (s *Server) handleBotPlanOrders(w http.ResponseWriter, r *http.Request, planID int64) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload struct {
		TelegramID string `json:"telegram_id"`
	}
	if err := decodeBotJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	order, err := s.botSalesService().CreateOrder(r.Context(), admin.ID, planID, payload.TelegramID)
	if err != nil {
		writeBotError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

// handleBotOrders lists recent orders of the selling admin (Gamaj Bot web
// management panel uses this to render the orders table).
func (s *Server) handleBotOrders(w http.ResponseWriter, r *http.Request) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	orders, err := s.botSalesService().Orders(r.Context(), admin.ID, limit)
	if err != nil {
		writeBotError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (s *Server) handleBotOrder(w http.ResponseWriter, r *http.Request, orderID string) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	order, err := s.botSalesService().Order(r.Context(), admin.ID, orderID)
	if err != nil {
		writeBotError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) handleBotWallet(w http.ResponseWriter, r *http.Request) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	buyer := strings.TrimSpace(r.URL.Query().Get("telegram_id"))
	if buyer == "" {
		writeError(w, http.StatusBadRequest, "telegram_id is required")
		return
	}
	switch r.Method {
	case http.MethodGet:
		balance, err := s.botSalesService().WalletBalance(r.Context(), admin.ID, buyer)
		if err != nil {
			writeBotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"balance": balance})
	case http.MethodPost:
		var payload struct {
			Amount    int64  `json:"amount"`
			Reference string `json:"reference"`
		}
		if err := decodeBotJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		balance, err := s.botSalesService().TopUpWallet(r.Context(), admin.ID, buyer, payload.Amount, strings.TrimSpace(payload.Reference))
		if err != nil {
			writeBotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"balance": balance})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleBotOrderAction(w http.ResponseWriter, r *http.Request, orderID, action string) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	switch action {
	case "pay-wallet":
		order, err := s.botSalesService().PayWithWallet(r.Context(), admin.ID, orderID)
		if err != nil {
			writeBotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, order)
	case "payment-link":
		var payload struct {
			CallbackURL string `json:"callback_url"`
		}
		_ = decodeBotJSON(r, &payload)
		callbackBase := strings.TrimSpace(payload.CallbackURL)
		if callbackBase == "" {
			if sales, err := s.settingsRepo.SalesSettings(r.Context()); err == nil && strings.TrimSpace(sales.TetraminatorCallback) != "" {
				callbackBase = strings.TrimSpace(sales.TetraminatorCallback)
			} else if err == nil && strings.TrimSpace(sales.TetraminatorBaseURL) == "" {
				callbackBase = requestOrigin(r)
			} else {
				callbackBase = requestOrigin(r)
			}
		}
		order, link, err := s.botSalesService().CreatePaymentLink(r.Context(), admin.ID, orderID, callbackBase+"/api/bot/payment/callback")
		if err != nil {
			writeBotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"order_id":    order.ID,
			"payment_url": link,
		})
	case "verify":
		var payload struct {
			PayID string `json:"pay_id"`
		}
		if err := decodeBotJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		order, err := s.botSalesService().VerifyGatewayPayment(r.Context(), admin.ID, orderID, payload.PayID)
		if err != nil {
			writeBotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, order)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// handleBotPaymentCallback receives the Tetraminator GET callback. It only
// records that the callback arrived; the actual settlement always runs via
// the inquiry endpoint through /verify.
func (s *Server) handleBotPaymentCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	payID := strings.TrimSpace(r.URL.Query().Get("pay_id"))
	if orderID == "" {
		orderID = strings.TrimSpace(r.URL.Query().Get("track_id"))
	}
	if err := s.botSalesService().HandleGatewayCallback(r.Context(), orderID, payID); err != nil {
		writeBotError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleBotService(w http.ResponseWriter, r *http.Request) {
	admin, err := s.botSalesPrincipal(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	buyer := strings.TrimSpace(r.URL.Query().Get("telegram_id"))
	if buyer == "" {
		writeError(w, http.StatusBadRequest, "telegram_id is required")
		return
	}
	view, err := s.botSalesService().BuyerService(r.Context(), admin.ID, buyer)
	if err != nil {
		writeBotError(w, err)
		return
	}
	// Enrich with the live subscription links from the user core.
	if detail, err := s.userService.UserGet(r.Context(), userapp.UserGetRequest{
		Username: view.Username,
		Admin: userapp.AdminContext{
			Username:       admin.Username,
			ID:             &admin.ID,
			Role:           string(admin.Role),
			CanViewTraffic: true,
			CanSortTraffic: true,
		},
	}); err == nil {
		view.Links = detail.Links
		view.SubscriptionURL = detail.SubscriptionURL
		view.Status = detail.Status
	}
	writeJSON(w, http.StatusOK, view)
}

// handleBotPlansPath dispatches POST /api/bot/plans/{plan_id}/orders.
func (s *Server) handleBotPlansPath(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/bot/plans/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[1] != "orders" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	planID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || planID <= 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	s.handleBotPlanOrders(w, r, planID)
}

// handleBotOrdersPath dispatches GET /api/bot/orders/{id} and
// POST /api/bot/orders/{id}/{action}.
func (s *Server) handleBotOrdersPath(w http.ResponseWriter, r *http.Request) {
	orderID, action, ok := botOrderIDFromPath(r.URL.Path, "/api/bot/orders/")
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if action == "" {
		s.handleBotOrder(w, r, orderID)
		return
	}
	s.handleBotOrderAction(w, r, orderID, action)
}

func botOrderIDFromPath(path, prefix string) (orderID, action string, ok bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return "", "", false
}
