package fraud

import (
	"testing"
	"time"
)

func TestTypesInitialization(t *testing.T) {
	req := PayoutRequest{
		RequestID:   "req_101",
		MerchantID:  "merch_99",
		AmountCents: 50000,
		Currency:    "GBP",
		CreatedAt:   time.Now(),
	}
	if req.RequestID != "req_101" {
		t.Fatalf("unexpected request ID: %s", req.RequestID)
	}
}
