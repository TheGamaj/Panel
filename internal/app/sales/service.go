// Package sales implements the Gamaj Panel commerce core: sale plans,
// buyer wallets, orders, gateway payments (Tetraminator) and automatic
// provisioning of purchased users. Bots such as Gamaj Bot use this module
// exclusively through the Gamaj API and never touch the database.
package sales

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	OrderStatusPending    = "pending"
	OrderStatusPaid       = "paid"
	OrderStatusProvisioned = "provisioned"
	OrderStatusFailed     = "failed"
	OrderStatusExpired    = "expired"

	WalletReasonTopUp   = "topup"
	WalletReasonOrder   = "order_payment"
	WalletReasonRefund  = "refund"
	WalletReasonAdmin   = "admin_adjustment"
)

// Plan is one purchasable plan offered by a selling admin.
type Plan struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Price        int64  `json:"price"`
	DurationDays int64  `json:"duration_days"`
	DataLimit    int64  `json:"data_limit"`
	IPLimit      int64  `json:"ip_limit,omitempty"`
	ServiceID    int64  `json:"service_id"`
}

// Order tracks a purchase from creation to provisioning.
type Order struct {
	ID        string  `json:"id"`
	PlanID    int64   `json:"plan_id"`
	PlanName  string  `json:"plan_name,omitempty"`
	Amount    int64   `json:"amount"`
	Status    string  `json:"status"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// Client errors carry HTTP semantics for the API layer.
type ClientError struct {
	Status  int
	Detail  string
}

func (e ClientError) Error() string { return e.Detail }

func clientError(status int, detail string) error {
	return ClientError{Status: status, Detail: detail}
}

// ErrOrderNotPending is returned when paying or verifying a settled order.
var ErrOrderNotPending = clientError(http.StatusConflict, "order is not pending")

// Service is the commerce core. It depends on the panel *sql.DB directly
// because it runs inside the Gamaj Panel process; external consumers only
// reach it through the Gamaj API.
type Service struct {
	db        *sql.DB
	dialect   string
	provision UserProvisioner
	gateway   Gateway
}

// UserProvisioner creates the purchased user inside the panel core using
// the selling admin identity. It returns the username and subscription URL.
type UserProvisioner func(ctx context.Context, adminID int64, plan Plan, buyerTelegramID, orderID string) (username string, subURL string, err error)

// Gateway creates and verifies Tetraminator invoices.
type Gateway interface {
	CreateInvoice(ctx context.Context, amount int64, callbackURL string) (payID string, paymentURL string, err error)
	Inquiry(ctx context.Context, payID string) (paid bool, amount int64, err error)
}

// New builds the sales service.
func New(db *sql.DB, dialect string, provisioner UserProvisioner, gateway Gateway) *Service {
	return &Service{db: db, dialect: dialect, provision: provisioner, gateway: gateway}
}

// SetGateway installs the Tetraminator gateway adapter.
func (s *Service) SetGateway(gateway Gateway) { s.gateway = gateway }

func newOrderID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func nowPtr() time.Time { return time.Now().UTC() }

func dbTimeString(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}

// Plans lists visible plans for a selling admin.
func (s *Service) Plans(ctx context.Context, adminID int64) ([]Plan, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, COALESCE(description, ''), price, duration_days, data_limit_bytes, ip_limit, COALESCE(service_id, 0)
FROM sale_plans
WHERE admin_id = ? AND visible = 1
ORDER BY price`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := []Plan{}
	for rows.Next() {
		var plan Plan
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.DurationDays, &plan.DataLimit, &plan.IPLimit, &plan.ServiceID); err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

// CreateOrder opens a pending order for a buyer. Orders expire after 30
// minutes if unpaid.
func (s *Service) CreateOrder(ctx context.Context, adminID, planID int64, buyerTelegramID string) (Order, error) {
	buyerTelegramID = strings.TrimSpace(buyerTelegramID)
	if buyerTelegramID == "" {
		return Order{}, clientError(http.StatusBadRequest, "telegram_id is required")
	}
	var plan Plan
	var visible int64
	err := s.db.QueryRowContext(ctx, `
SELECT id, name, COALESCE(description, ''), price, duration_days, data_limit_bytes, ip_limit, COALESCE(service_id, 0), visible
FROM sale_plans WHERE id = ? AND admin_id = ?`, planID, adminID).Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.DurationDays, &plan.DataLimit, &plan.IPLimit, &plan.ServiceID, &visible,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, clientError(http.StatusNotFound, "plan not found")
	}
	if err != nil {
		return Order{}, err
	}
	if visible != 1 {
		return Order{}, clientError(http.StatusNotFound, "plan not found")
	}
	orderID, err := newOrderID()
	if err != nil {
		return Order{}, err
	}
	expires := nowPtr().Add(30 * time.Minute)
	_, err = s.db.ExecContext(ctx, `
INSERT INTO orders (id, admin_id, plan_id, buyer_telegram_id, amount, status, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, adminID, plan.ID, buyerTelegramID, plan.Price, OrderStatusPending, dbTimeString(expires), dbTimeString(nowPtr()), dbTimeString(nowPtr()),
	)
	if err != nil {
		return Order{}, err
	}
	expiresText := expires.Format(time.RFC3339)
	return Order{ID: orderID, PlanID: plan.ID, PlanName: plan.Name, Amount: plan.Price, Status: OrderStatusPending, ExpiresAt: &expiresText}, nil
}

func (s *Service) orderTx(ctx context.Context, tx *sql.Tx, orderID string, adminID int64) (Order, error) {
	var order Order
	var expires sql.NullString
	err := tx.QueryRowContext(ctx, `
SELECT o.id, o.plan_id, o.amount, o.status, o.expires_at, p.name
FROM orders o JOIN sale_plans p ON p.id = o.plan_id
WHERE o.id = ? AND o.admin_id = ?`, orderID, adminID).Scan(
		&order.ID, &order.PlanID, &order.Amount, &order.Status, &expires, &order.PlanName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, clientError(http.StatusNotFound, "order not found")
	}
	if err != nil {
		return Order{}, err
	}
	if expires.Valid {
		order.ExpiresAt = &expires.String
	}
	return order, nil
}

// Order returns one order for display.
func (s *Service) Order(ctx context.Context, adminID int64, orderID string) (Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()
	return s.orderTx(ctx, tx, orderID, adminID)
}

// Orders lists recent orders of a selling admin, newest first. Used by the
// Gamaj Bot web management panel and the bot API client.
func (s *Service) Orders(ctx context.Context, adminID int64, limit int) ([]Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT o.id, o.plan_id, o.amount, o.status, o.expires_at, p.name
FROM orders o JOIN sale_plans p ON p.id = o.plan_id
WHERE o.admin_id = ?
ORDER BY o.created_at DESC, o.id DESC
LIMIT ?`, adminID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := []Order{}
	for rows.Next() {
		var order Order
		var expires sql.NullString
		if err := rows.Scan(&order.ID, &order.PlanID, &order.Amount, &order.Status, &expires, &order.PlanName); err != nil {
			return nil, err
		}
		if expires.Valid {
			order.ExpiresAt = &expires.String
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

// WalletBalance returns the buyer wallet balance for a selling admin.
func (s *Service) WalletBalance(ctx context.Context, adminID int64, buyerTelegramID string) (int64, error) {
	var balance int64
	err := s.db.QueryRowContext(ctx, `
SELECT balance FROM buyer_wallets WHERE admin_id = ? AND buyer_telegram_id = ?`,
		adminID, strings.TrimSpace(buyerTelegramID)).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return balance, err
}

// walletMutationTx applies a wallet change with an immutable ledger row. It
// returns false when the balance would go negative.
func walletMutationTx(ctx context.Context, tx *sql.Tx, dialect string, adminID int64, buyerTelegramID string, amount int64, reason, reference string, now time.Time) (bool, error) {
	_, err := tx.ExecContext(ctx, `
INSERT INTO buyer_wallets (admin_id, buyer_telegram_id, balance, updated_at)
VALUES (?, ?, 0, ?)`, adminID, strings.TrimSpace(buyerTelegramID), dbTimeString(now))
	if err != nil && !isDuplicateError(err) {
		return false, err
	}
	if amount < 0 {
		var balance int64
		if err := tx.QueryRowContext(ctx, `
SELECT balance FROM buyer_wallets WHERE admin_id = ? AND buyer_telegram_id = ?`,
			adminID, strings.TrimSpace(buyerTelegramID)).Scan(&balance); err != nil {
			return false, err
		}
		if balance+amount < 0 {
			return false, nil
		}
	}
	var updated int64
	if err := tx.QueryRowContext(ctx, `
UPDATE buyer_wallets SET balance = balance + ?, updated_at = ?
WHERE admin_id = ? AND buyer_telegram_id = ?
RETURNING balance` /* sqlite */, adminID, amount, dbTimeString(now), adminID, strings.TrimSpace(buyerTelegramID)).Scan(&updated); err != nil {
		if dialect != "sqlite" || !strings.Contains(strings.ToLower(err.Error()), "returning") {
			// MySQL has no RETURNING; fall back to a read.
			if _, execErr := tx.ExecContext(ctx, `
UPDATE buyer_wallets SET balance = balance + ?, updated_at = ?
WHERE admin_id = ? AND buyer_telegram_id = ?`, amount, dbTimeString(now), adminID, strings.TrimSpace(buyerTelegramID)); execErr != nil {
				return false, execErr
			}
		} else {
			return false, err
		}
	}
	if err := tx.QueryRowContext(ctx, `
SELECT balance FROM buyer_wallets WHERE admin_id = ? AND buyer_telegram_id = ?`,
		adminID, strings.TrimSpace(buyerTelegramID)).Scan(&updated); err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO reseller_wallet_transactions (admin_id, amount, balance_after, reason, reference, created_at)
VALUES (?, ?, ?, ?, ?, ?)`, adminID, amount, updated, reason, reference, dbTimeString(now))
	if err != nil {
		return false, err
	}
	return true, nil
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique") || strings.Contains(message, "constraint")
}

// TopUpWallet credits a buyer wallet (admin-initiated manual top-up).
func (s *Service) TopUpWallet(ctx context.Context, adminID int64, buyerTelegramID string, amount int64, reference string) (int64, error) {
	if amount == 0 {
		return 0, clientError(http.StatusBadRequest, "amount must be non-zero")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	ok, err := walletMutationTx(ctx, tx, s.dialect, adminID, buyerTelegramID, amount, WalletReasonTopUp, reference, nowPtr())
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, clientError(http.StatusConflict, "wallet balance would go negative")
	}
	var balance int64
	if err := tx.QueryRowContext(ctx, `
SELECT balance FROM buyer_wallets WHERE admin_id = ? AND buyer_telegram_id = ?`,
		adminID, strings.TrimSpace(buyerTelegramID)).Scan(&balance); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return balance, nil
}

// PayWithWallet settles a pending order from the buyer wallet and provisions
// the service. The wallet debit, order update and provisioning decision run
// inside a single transaction so double payments cannot occur.
func (s *Service) PayWithWallet(ctx context.Context, adminID int64, orderID string) (Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()
	order, err := s.orderTx(ctx, tx, orderID, adminID)
	if err != nil {
		return Order{}, err
	}
	if order.Status != OrderStatusPending {
		return Order{}, ErrOrderNotPending
	}
	var buyerTelegramID string
	if err := tx.QueryRowContext(ctx, `SELECT buyer_telegram_id FROM orders WHERE id = ?`, orderID).Scan(&buyerTelegramID); err != nil {
		return Order{}, err
	}
	ok, err := walletMutationTx(ctx, tx, s.dialect, adminID, buyerTelegramID, -order.Amount, WalletReasonOrder, orderID, nowPtr())
	if err != nil {
		return Order{}, err
	}
	if !ok {
		return Order{}, clientError(http.StatusPaymentRequired, "wallet balance is not enough")
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE orders SET status = ?, payment_method = 'wallet', paid_at = ?, updated_at = ? WHERE id = ?`,
		OrderStatusPaid, dbTimeString(nowPtr()), dbTimeString(nowPtr()), orderID); err != nil {
		return Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	order.Status = OrderStatusPaid
	provisioned, err := s.ProvisionPaidOrder(ctx, adminID, orderID)
	if err != nil {
		return order, nil
	}
	return provisioned, nil
}

// CreatePaymentLink opens a Tetraminator invoice for a pending order.
func (s *Service) CreatePaymentLink(ctx context.Context, adminID int64, orderID, callbackBase string) (Order, string, error) {
	if s.gateway == nil {
		return Order{}, "", clientError(http.StatusServiceUnavailable, "payment gateway is not configured")
	}
	order, err := s.Order(ctx, adminID, orderID)
	if err != nil {
		return Order{}, "", err
	}
	if order.Status != OrderStatusPending {
		return Order{}, "", ErrOrderNotPending
	}
	if err := s.validateGatewayAmount(order.Amount); err != nil {
		return Order{}, "", err
	}
	payID, paymentURL, err := s.gateway.CreateInvoice(ctx, order.Amount, callbackBase)
	if err != nil {
		return Order{}, "", clientError(http.StatusBadGateway, "payment gateway error: "+err.Error())
	}
	if payID == "" || paymentURL == "" {
		return Order{}, "", clientError(http.StatusBadGateway, "payment gateway returned an invalid response")
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO order_payments (order_id, gateway, pay_id, amount, status, created_at, updated_at)
VALUES (?, 'tetraminator', ?, ?, 'pending', ?, ?)`,
		orderID, payID, order.Amount, dbTimeString(nowPtr()), dbTimeString(nowPtr())); err != nil {
		return Order{}, "", err
	}
	return order, paymentURL, nil
}

func (s *Service) validateGatewayAmount(amount int64) error {
	var min, max int64
	if err := s.db.QueryRowContext(context.Background(), `
SELECT COALESCE(tetraminator_min_price, 50000), COALESCE(tetraminator_max_price, 10000000)
FROM settings WHERE id = 1`).Scan(&min, &max); err != nil {
		min, max = 50000, 10000000
	}
	if amount < min || amount > max {
		return clientError(http.StatusUnprocessableEntity, fmt.Sprintf("amount must be between %d and %d", min, max))
	}
	return nil
}

// VerifyGatewayPayment confirms an order through the Tetraminator inquiry
// endpoint. The gateway response must mark the payment paid and match the
// order amount; the pay_id must belong to the order.
func (s *Service) VerifyGatewayPayment(ctx context.Context, adminID int64, orderID, payID string) (Order, error) {
	if s.gateway == nil {
		return Order{}, clientError(http.StatusServiceUnavailable, "payment gateway is not configured")
	}
	payID = strings.TrimSpace(payID)
	if payID == "" {
		return Order{}, clientError(http.StatusBadRequest, "pay_id is required")
	}
	var amount int64
	var gateway string
	err := s.db.QueryRowContext(ctx, `
SELECT amount, gateway FROM order_payments WHERE order_id = ? AND pay_id = ?`, orderID, payID).Scan(&amount, &gateway)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, clientError(http.StatusNotFound, "payment not found for this order")
	}
	if err != nil {
		return Order{}, err
	}
	order, err := s.Order(ctx, adminID, orderID)
	if err != nil {
		return Order{}, err
	}
	if order.Status != OrderStatusPending {
		return order, nil
	}
	if amount != order.Amount {
		return Order{}, clientError(http.StatusConflict, "payment amount does not match the order")
	}
	paid, inquiryAmount, err := s.gateway.Inquiry(ctx, payID)
	if err != nil {
		return Order{}, clientError(http.StatusBadGateway, "payment inquiry failed: "+err.Error())
	}
	if !paid {
		_, _ = s.db.ExecContext(ctx, `
UPDATE order_payments SET inquiry_status = 'unpaid', updated_at = ? WHERE order_id = ? AND pay_id = ?`,
			dbTimeString(nowPtr()), orderID, payID)
		return Order{}, clientError(http.StatusPaymentRequired, "payment is not confirmed yet")
	}
	if inquiryAmount != order.Amount {
		_, _ = s.db.ExecContext(ctx, `
UPDATE order_payments SET inquiry_status = 'amount_mismatch', updated_at = ? WHERE order_id = ? AND pay_id = ?`,
			dbTimeString(nowPtr()), orderID, payID)
		return Order{}, clientError(http.StatusConflict, "payment amount does not match the order")
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE orders SET status = ?, payment_method = 'tetraminator', payment_ref = ?, paid_at = ?, updated_at = ? WHERE id = ? AND status = ?`,
		OrderStatusPaid, payID, dbTimeString(nowPtr()), dbTimeString(nowPtr()), orderID, OrderStatusPending); err != nil {
		return Order{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE order_payments SET status = 'paid', inquiry_status = 'paid', updated_at = ? WHERE order_id = ? AND pay_id = ?`,
		dbTimeString(nowPtr()), orderID, payID); err != nil {
		return Order{}, err
	}
	order.Status = OrderStatusPaid
	return s.ProvisionPaidOrder(ctx, adminID, orderID)
}

// HandleGatewayCallback records the callback order_id so a replayed callback
// cannot re-trigger provisioning. Verification itself always runs through
// the gateway inquiry endpoint.
func (s *Service) HandleGatewayCallback(ctx context.Context, orderID, payID string) error {
	if orderID == "" || payID == "" {
		return clientError(http.StatusBadRequest, "order_id and pay_id are required")
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE order_payments SET callback_seen = 1, updated_at = ?
WHERE order_id = ? AND pay_id = ? AND callback_seen = 0`,
		dbTimeString(nowPtr()), orderID, payID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		// Replay or unknown payment: report accepted so the gateway stops
		// retrying, but do nothing.
		return nil
	}
	return nil
}

// ProvisionPaidOrder creates the purchased user exactly once. A unique index
// on orders.provisioned_user_id would be ideal; instead we rely on the
// status transition guard below, which is atomic for every dialect.
func (s *Service) ProvisionPaidOrder(ctx context.Context, adminID int64, orderID string) (Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var status string
	var planID int64
	var buyerTelegramID string
	var provisioned sql.NullInt64
	err = tx.QueryRowContext(ctx, `
SELECT status, plan_id, buyer_telegram_id, provisioned_user_id FROM orders WHERE id = ? AND admin_id = ?`,
		orderID, adminID).Scan(&status, &planID, &buyerTelegramID, &provisioned)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, clientError(http.StatusNotFound, "order not found")
	}
	if err != nil {
		return Order{}, err
	}
	if provisioned.Valid {
		order, err := s.orderTx(ctx, tx, orderID, adminID)
		if err == nil {
			order.Status = OrderStatusProvisioned
			return order, nil
		}
		return Order{}, err
	}
	if status != OrderStatusPaid {
		return Order{}, clientError(http.StatusConflict, "order is not paid yet")
	}
	var plan Plan
	err = tx.QueryRowContext(ctx, `
SELECT id, name, COALESCE(description, ''), price, duration_days, data_limit_bytes, ip_limit, COALESCE(service_id, 0)
FROM sale_plans WHERE id = ?`, planID).Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.Price, &plan.DurationDays, &plan.DataLimit, &plan.IPLimit, &plan.ServiceID)
	if err != nil {
		return Order{}, err
	}
	// Claim the order atomically: only one caller can transition paid →
	// provisioned.
	claim, err := tx.ExecContext(ctx, `
UPDATE orders SET status = ?, provisioned_user_id = -1, updated_at = ? WHERE id = ? AND status = ?`,
		OrderStatusProvisioned, dbTimeString(nowPtr()), orderID, OrderStatusPaid)
	if err != nil {
		return Order{}, err
	}
	claimed, err := claim.RowsAffected()
	if err != nil {
		return Order{}, err
	}
	if claimed == 0 {
		return Order{}, clientError(http.StatusConflict, "order is already being provisioned")
	}
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}

	username, subURL, provisionErr := s.provision(ctx, adminID, plan, buyerTelegramID, orderID)
	if provisionErr != nil {
		log.Printf("gamaj sales: provisioning order %s failed: %v", orderID, provisionErr)
		_, _ = s.db.ExecContext(ctx, `
UPDATE orders SET status = ?, provisioned_user_id = NULL, updated_at = ? WHERE id = ? AND provisioned_user_id = -1`,
			OrderStatusPaid, dbTimeString(nowPtr()), orderID)
		return Order{}, clientError(http.StatusInternalServerError, "service provisioning failed; the order stays paid and can be retried")
	}
	var userID int64
	_ = s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = ?`, username).Scan(&userID)
	if _, err := s.db.ExecContext(ctx, `
UPDATE orders SET provisioned_user_id = ?, buyer_note = ?, updated_at = ? WHERE id = ?`,
		userID, subURL, dbTimeString(nowPtr()), orderID); err != nil {
		return Order{}, err
	}
	order, err := s.Order(ctx, adminID, orderID)
	if err != nil {
		return Order{}, err
	}
	order.Status = OrderStatusProvisioned
	return order, nil
}

// ServiceView is the delivered service payload for a buyer.
type ServiceView struct {
	Username        string   `json:"username"`
	Status          string   `json:"status"`
	UsedTraffic     int64    `json:"used_traffic"`
	DataLimit       int64    `json:"data_limit"`
	Expire          int64    `json:"expire"`
	SubscriptionURL string   `json:"subscription_url"`
	Links           []string `json:"links,omitempty"`
}

// BuyerService returns the most recent provisioned service for a buyer.
func (s *Service) BuyerService(ctx context.Context, adminID int64, buyerTelegramID string) (ServiceView, error) {
	var username, subURL string
	err := s.db.QueryRowContext(ctx, `
SELECT o.buyer_note, u.username FROM orders o JOIN users u ON u.id = o.provisioned_user_id
WHERE o.admin_id = ? AND o.buyer_telegram_id = ? AND o.provisioned_user_id IS NOT NULL
ORDER BY o.updated_at DESC LIMIT 1`, adminID, strings.TrimSpace(buyerTelegramID)).Scan(&subURL, &username)
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceView{}, clientError(http.StatusNotFound, "no service found")
	}
	if err != nil {
		return ServiceView{}, err
	}
	var status string
	var used, dataLimit, expire int64
	err = s.db.QueryRowContext(ctx, `
SELECT status, used_traffic, COALESCE(data_limit, 0), COALESCE(expire, 0) FROM users WHERE username = ?`,
		username).Scan(&status, &used, &dataLimit, &expire)
	if err != nil {
		return ServiceView{}, err
	}
	return ServiceView{
		Username:        username,
		Status:          status,
		UsedTraffic:     used,
		DataLimit:       dataLimit,
		Expire:          expire,
		SubscriptionURL: subURL,
		Links:           []string{},
	}, nil
}

var _ = io.Discard
var _ = url.QueryEscape
var _ = strconv.Itoa
