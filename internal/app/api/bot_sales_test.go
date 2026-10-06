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

// TestSalesPlanManagement covers what a management panel does to a plan:
// read them all including retired ones, edit one field without disturbing the
// rest, retire one, and bring it back. Retiring is deliberately not a delete,
// and the order history assertion is what pins that down.
func TestSalesPlanManagement(t *testing.T) {
	database, dialect := openSalesTestDB(t)
	adminID, serviceID, planID := insertSalesFixtures(t, database)
	service := salesapp.New(database, dialect, nil, nil)
	ctx := context.Background()

	// A plan belonging to another reseller must be unreachable even when the
	// caller holds its id. Scoping has to happen in the query, not in a check
	// after the read.
	other, err := database.Exec(`INSERT INTO admins (username, hashed_password, role, permissions, status) VALUES ('other', 'x', 'full_access', '{}', 'active')`)
	if err != nil {
		t.Fatalf("seed other admin: %v", err)
	}
	otherAdminID, _ := other.LastInsertId()
	otherInsert, err := database.Exec(`INSERT INTO sale_plans (admin_id, name, price, duration_days, data_limit_bytes, service_id, visible) VALUES (?, 'theirs', 5000, 30, 0, ?, 1)`, otherAdminID, serviceID)
	if err != nil {
		t.Fatalf("seed other plan: %v", err)
	}
	otherPlanID, _ := otherInsert.LastInsertId()

	price := int64(150000)
	if _, err := service.UpdatePlan(ctx, adminID, otherPlanID, salesapp.PlanUpdate{Price: &price}); err == nil {
		t.Fatal("updated another reseller's plan")
	}
	if err := service.RetirePlan(ctx, adminID, otherPlanID); err == nil {
		t.Fatal("retired another reseller's plan")
	}

	// A partial edit touches only what it names.
	updated, err := service.UpdatePlan(ctx, adminID, planID, salesapp.PlanUpdate{Price: &price})
	if err != nil {
		t.Fatalf("update price: %v", err)
	}
	if updated.Price != 150000 {
		t.Fatalf("price = %d, want 150000", updated.Price)
	}
	if updated.Name != "10GB" || updated.DurationDays != 30 || updated.DataLimit != 10737418240 || updated.ServiceID != serviceID {
		t.Fatalf("a partial edit changed untouched fields: %+v", updated.Plan)
	}
	if !updated.Visible {
		t.Fatal("a partial edit retired the plan")
	}

	// Validation runs against the merged plan, so a patch that only clears the
	// name is still rejected.
	blank := "   "
	if _, err := service.UpdatePlan(ctx, adminID, planID, salesapp.PlanUpdate{Name: &blank}); err == nil {
		t.Fatal("accepted a blank name")
	}
	zero := int64(0)
	if _, err := service.UpdatePlan(ctx, adminID, planID, salesapp.PlanUpdate{Price: &zero}); err == nil {
		t.Fatal("accepted a zero price")
	}
	if _, err := service.UpdatePlan(ctx, adminID, planID, salesapp.PlanUpdate{ServiceID: &zero}); err == nil {
		t.Fatal("accepted service_id 0")
	}
	// The rejected patches must not have been written.
	after, err := service.ManagePlans(ctx, adminID)
	if err != nil || len(after) != 1 {
		t.Fatalf("re-read plans: %d records, %v", len(after), err)
	}
	if after[0].Price != 150000 || after[0].Name != "10GB" || after[0].ServiceID != serviceID {
		t.Fatalf("a rejected patch was written anyway: %+v", after[0].Plan)
	}

	// An order exists before the plan is retired, so the orders query keeps a
	// join to the plans table. A hard delete would drop this history.
	if _, err := service.CreateOrder(ctx, adminID, planID, "12345"); err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := service.RetirePlan(ctx, adminID, planID); err != nil {
		t.Fatalf("retire: %v", err)
	}
	// Retiring twice must not read as "not found": an update that changes
	// nothing reports zero affected rows on MySQL, so existence is checked
	// before the write rather than inferred from it.
	if err := service.RetirePlan(ctx, adminID, planID); err != nil {
		t.Fatalf("retire again: %v", err)
	}

	// The buyer-facing list hides it; the management list still shows it.
	visible, err := service.Plans(ctx, adminID)
	if err != nil {
		t.Fatalf("plans: %v", err)
	}
	if len(visible) != 0 {
		t.Fatalf("buyer list returned %d plans, want 0", len(visible))
	}
	managed, err := service.ManagePlans(ctx, adminID)
	if err != nil {
		t.Fatalf("manage plans: %v", err)
	}
	if len(managed) != 1 || managed[0].ID != planID || managed[0].Visible {
		t.Fatalf("management list = %+v, want the one retired plan", managed)
	}
	for _, record := range managed {
		if record.ID == otherPlanID {
			t.Fatal("management list leaked another reseller's plan")
		}
	}

	// The order history survived the retirement.
	orders, err := service.Orders(ctx, adminID, 0)
	if err != nil {
		t.Fatalf("orders: %v", err)
	}
	if len(orders) != 1 || orders[0].PlanName != "10GB" {
		t.Fatalf("orders = %+v, want the order kept with its plan name", orders)
	}

	// And a retired plan can be put back on sale.
	on := true
	restored, err := service.UpdatePlan(ctx, adminID, planID, salesapp.PlanUpdate{Visible: &on})
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restored.Visible {
		t.Fatal("plan did not come back on sale")
	}
	visible, err = service.Plans(ctx, adminID)
	if err != nil || len(visible) != 1 {
		t.Fatalf("buyer list after restore = %d plans, %v", len(visible), err)
	}
}

// TestBotRoutesAreRegistered proves the /api/bot surface is genuinely wired
// into the router. Every handler above can exist while answering 404, which is
// exactly what happened before: the whole file was unrouted. The previous
// version of this test could not catch that, because it wrapped a stub handler
// with requireAdmin and never touched the router at all.
//
// The server is deliberately bare: an anonymous request is rejected by
// requireAdmin before any handler runs, so these assertions need no database.
// If the routes are ever unregistered again the request falls through to the
// subscription 404 handler, which panics on the nil database — still a hard
// failure, just a noisier one.
func TestBotRoutesAreRegistered(t *testing.T) {
	handler := (&Server{}).Handler()

	// Every sales route is authenticated by the bot's API key, so an anonymous
	// request must be rejected as unauthorized — never as "not found". This
	// stops before any database access, because credentials are resolved first.
	salesRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/bot/plans"},
		{http.MethodPost, "/api/bot/plans"},
		{http.MethodPut, "/api/bot/plans/7"},
		{http.MethodPatch, "/api/bot/plans/7"},
		{http.MethodDelete, "/api/bot/plans/7"},
		{http.MethodPost, "/api/bot/plans/7/orders"},
		{http.MethodGet, "/api/bot/orders"},
		{http.MethodGet, "/api/bot/orders/order-1"},
		{http.MethodPost, "/api/bot/orders/order-1/pay-wallet"},
		{http.MethodPost, "/api/bot/orders/order-1/payment-link"},
		{http.MethodPost, "/api/bot/orders/order-1/verify"},
		{http.MethodGet, "/api/bot/wallet?telegram_id=1"},
		{http.MethodGet, "/api/bot/service?telegram_id=1"},
	}
	for _, route := range salesRoutes {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(route.method, route.path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401: the route is missing or unguarded", route.method, route.path, recorder.Code)
		}
	}

	// The gateway callback is unauthenticated on purpose: Tetraminator calls it
	// back with no Gamaj credential. It only answers GET, so a POST settles the
	// response without reaching the payment service.
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/bot/payment/callback", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/bot/payment/callback status = %d, want 405", recorder.Code)
	}

	// The 401s above cannot pass by accident: requireAdmin is only reachable
	// once the router has matched a route, and an unmatched path never produces
	// 401 — it falls through to the 404 handler instead.
}
