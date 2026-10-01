package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheGamaj/Panel/internal/app/migrations"
	salesapp "github.com/TheGamaj/Panel/internal/app/sales"
	_ "modernc.org/sqlite"
)

// stubGateway is a deterministic Tetraminator double for tests.
type stubGateway struct {
	invoicePayID  string
	invoiceURL    string
	inquiryPaid   bool
	inquiryAmount int64
	createdAmount int64
}

func (g *stubGateway) CreateInvoice(ctx context.Context, amount int64, callbackURL string) (string, string, error) {
	g.createdAmount = amount
	return g.invoicePayID, g.invoiceURL, nil
}

func (g *stubGateway) Inquiry(ctx context.Context, payID string) (bool, int64, error) {
	return g.inquiryPaid, g.inquiryAmount, nil
}

func openSalesTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	database, err := sql.Open("sqlite", filepath.ToSlash(filepath.Join(t.TempDir(), "sales.sqlite3")))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := migrations.RunMigrations(ctx, database, "sqlite"); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	return database, "sqlite"
}

func insertSalesFixtures(t *testing.T, database *sql.DB) (adminID, serviceID, planID int64) {
	t.Helper()
	result, err := database.Exec(`INSERT INTO admins (username, hashed_password, role, permissions, status) VALUES ('seller', 'x', 'full_access', '{}', 'active')`)
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	adminID, _ = result.LastInsertId()
	result, err = database.Exec(`INSERT INTO services (name) VALUES ('Gamaj Service')`)
	if err != nil {
		t.Fatalf("seed service: %v", err)
	}
	serviceID, _ = result.LastInsertId()
	result, err = database.Exec(`INSERT INTO sale_plans (admin_id, name, price, duration_days, data_limit_bytes, service_id, visible) VALUES (?, '10GB', 100000, 30, 10737418240, ?, 1)`, adminID, serviceID)
	if err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	planID, _ = result.LastInsertId()
	return adminID, serviceID, planID
}

func TestSalesWalletOrderFlow(t *testing.T) {
	database, dialect := openSalesTestDB(t)
	adminID, _, planID := insertSalesFixtures(t, database)

	var provisioned int64
	service := salesapp.New(database, dialect, func(ctx context.Context, admin int64, plan salesapp.Plan, buyer, order string) (string, string, error) {
		provisioned++
		return "bot-testuser", "https://panel.example/sub/bot-testuser", nil
	}, nil)

	// Wallet top-up then wallet payment.
	if _, err := service.TopUpWallet(context.Background(), adminID, "12345", 150000, "seed"); err != nil {
		t.Fatalf("top up: %v", err)
	}
	balance, err := service.WalletBalance(context.Background(), adminID, "12345")
	if err != nil || balance != 150000 {
		t.Fatalf("wallet balance = %d, %v", balance, err)
	}
	order, err := service.CreateOrder(context.Background(), adminID, planID, "12345")
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	paid, err := service.PayWithWallet(context.Background(), adminID, order.ID)
	if err != nil {
		t.Fatalf("pay with wallet: %v", err)
	}
	if paid.Status != salesapp.OrderStatusProvisioned {
		t.Fatalf("order status = %s, want provisioned", paid.Status)
	}
	if provisioned != 1 {
		t.Fatalf("provisioner called %d times, want 1", provisioned)
	}
	balance, err = service.WalletBalance(context.Background(), adminID, "12345")
	if err != nil || balance != 50000 {
		t.Fatalf("wallet after payment = %d, %v", balance, err)
	}

	// Insufficient balance must fail and keep the order pending.
	order2, err := service.CreateOrder(context.Background(), adminID, planID, "12345")
	if err != nil {
		t.Fatalf("create order 2: %v", err)
	}
	if _, err := service.PayWithWallet(context.Background(), adminID, order2.ID); err == nil {
		t.Fatal("expected insufficient balance error")
	}
	final, err := service.Order(context.Background(), adminID, order2.ID)
	if err != nil || final.Status != salesapp.OrderStatusPending {
		t.Fatalf("order 2 status = %s, %v", final.Status, err)
	}
}

func TestSalesTetraminatorVerify(t *testing.T) {
	database, dialect := openSalesTestDB(t)
	adminID, _, planID := insertSalesFixtures(t, database)

	gateway := &stubGateway{invoicePayID: "pay-1", invoiceURL: "https://pay.example/1"}
	service := salesapp.New(database, dialect, func(ctx context.Context, admin int64, plan salesapp.Plan, buyer, order string) (string, string, error) {
		return "bot-verify", "https://panel.example/sub/bot-verify", nil
	}, gateway)
	service.SetGateway(gateway)

	order, err := service.CreateOrder(context.Background(), adminID, planID, "67890")
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	_, link, err := service.CreatePaymentLink(context.Background(), adminID, order.ID, "https://panel.example")
	if err != nil {
		t.Fatalf("payment link: %v", err)
	}
	if link != gateway.invoiceURL {
		t.Fatalf("payment link = %s", link)
	}

	// Unpaid inquiry must not provision.
	gateway.inquiryPaid = false
	if _, err := service.VerifyGatewayPayment(context.Background(), adminID, order.ID, "pay-1"); err == nil {
		t.Fatal("expected unpaid inquiry error")
	}

	// Amount mismatch must not provision.
	gateway.inquiryPaid = true
	gateway.inquiryAmount = 1
	if _, err := service.VerifyGatewayPayment(context.Background(), adminID, order.ID, "pay-1"); err == nil {
		t.Fatal("expected amount mismatch error")
	}

	// Correct amount settles and provisions exactly once.
	gateway.inquiryAmount = 100000
	paid, err := service.VerifyGatewayPayment(context.Background(), adminID, order.ID, "pay-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if paid.Status != salesapp.OrderStatusProvisioned {
		t.Fatalf("status = %s", paid.Status)
	}
	// Verifying again returns the settled order without a new provision.
	again, err := service.VerifyGatewayPayment(context.Background(), adminID, order.ID, "pay-1")
	if err != nil || again.Status != salesapp.OrderStatusProvisioned {
		t.Fatalf("re-verify: %s, %v", again.Status, err)
	}
}

func TestBotSalesEndpointsAuthorization(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	server := &Server{}
	// /bot routes exist and require authentication (no credentials => 401).
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/bot/plans", nil)
	server.requireAdmin(handler)(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated /bot/plans status = %d", recorder.Code)
	}
}
