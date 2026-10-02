package sales

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TetraminatorConfig carries the admin-configured gateway credentials.
type TetraminatorConfig struct {
	BaseURL string
	APIKey  string
}

// tetraminatorInvoiceResponse models POST /invoice/create.
type tetraminatorInvoiceResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	PayID       string `json:"pay_id"`
	PaymentLink string `json:"payment_link"`
}

// tetraminatorInquiryResponse models GET /payment/inquiry/{pay_id}.
type tetraminatorInquiryResponse struct {
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	PayID         string `json:"pay_id"`
	Amount        int64  `json:"amount"`
}

// Tetraminator is the payment gateway adapter. It validates every response
// shape before trusting it and never logs the API key.
type Tetraminator struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewTetraminator builds the adapter from admin settings.
func NewTetraminator(config TetraminatorConfig) *Tetraminator {
	return &Tetraminator{
		baseURL: strings.TrimRight(strings.TrimSpace(config.BaseURL), "/"),
		apiKey:  strings.TrimSpace(config.APIKey),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Valid reports whether the adapter has usable credentials.
func (t *Tetraminator) Valid() bool {
	return t != nil && t.baseURL != "" && t.apiKey != ""
}

func (t *Tetraminator) do(ctx context.Context, method, path string, payload any, target any) error {
	endpoint := t.baseURL + path
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = strings.NewReader(string(encoded))
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", t.apiKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := t.http.Do(req)
	if err != nil {
		return fmt.Errorf("gateway request failed")
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read gateway response failed")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("gateway returned HTTP %d", res.StatusCode)
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("gateway response is not valid JSON")
	}
	return nil
}

// CreateInvoice opens an invoice and returns (pay_id, payment_link).
func (t *Tetraminator) CreateInvoice(ctx context.Context, amount int64, callbackURL string) (string, string, error) {
	if !t.Valid() {
		return "", "", fmt.Errorf("gateway is not configured")
	}
	payload := map[string]any{
		"price":        amount,
		"callback_url": callbackURL,
	}
	var response tetraminatorInvoiceResponse
	if err := t.do(ctx, http.MethodPost, "/invoice/create", payload, &response); err != nil {
		return "", "", err
	}
	if !strings.EqualFold(response.Status, "success") && response.Message != "" && response.PayID == "" {
		return "", "", fmt.Errorf("gateway rejected the invoice: %s", response.Message)
	}
	if response.PayID == "" || response.PaymentLink == "" {
		return "", "", fmt.Errorf("gateway response is missing pay_id or payment_link")
	}
	if _, err := url.Parse(response.PaymentLink); err != nil {
		return "", "", fmt.Errorf("gateway returned an invalid payment link")
	}
	return response.PayID, response.PaymentLink, nil
}

// Inquiry asks the gateway whether a payment is confirmed.
func (t *Tetraminator) Inquiry(ctx context.Context, payID string) (bool, int64, error) {
	if !t.Valid() {
		return false, 0, fmt.Errorf("gateway is not configured")
	}
	var response tetraminatorInquiryResponse
	path := "/payment/inquiry/" + url.PathEscape(payID)
	if err := t.do(ctx, http.MethodGet, path, nil, &response); err != nil {
		return false, 0, err
	}
	if response.PayID == "" {
		return false, 0, fmt.Errorf("gateway inquiry response is missing pay_id")
	}
	return strings.EqualFold(response.PaymentStatus, "paid"), response.Amount, nil
}
