package fraud

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFraudEngineRules(t *testing.T) {
	rules := []Rule{
		&VelocityRule{WindowDuration: 1 * time.Hour, MaxCount: 3, MaxTotalCents: 100000},
		&AccountTamperRule{MinCooldownHours: 24},
		&AmountSpikeRule{MaxMultiplier: 3.0},
	}
	engine := NewEngine(rules, 70, 35)

	baseTime := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)

	t.Run("ApprovesNormalPayout", func(t *testing.T) {
		req := PayoutRequest{
			RequestID:       "req_norm",
			MerchantID:      "merch_1",
			AmountCents:     25000,
			Currency:        "GBP",
			BankAccountHash: "bank_abc",
			DeviceID:        "dev_1",
			AccountAgeHours: 72,
			CreatedAt:       baseTime,
		}
		history := []PayoutRequest{
			{RequestID: "h1", AmountCents: 20000, DeviceID: "dev_1", CreatedAt: baseTime.Add(-2 * time.Hour)},
			{RequestID: "h2", AmountCents: 22000, DeviceID: "dev_1", CreatedAt: baseTime.Add(-5 * time.Hour)},
		}

		res, err := engine.Evaluate(context.Background(), req, history)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Decision != DecisionApprove {
			t.Errorf("expected APPROVE, got %s (score: %d)", res.Decision, res.TotalScore)
		}
	})

	t.Run("ReviewsSpikeAndTamper", func(t *testing.T) {
		req := PayoutRequest{
			RequestID:       "req_spike",
			MerchantID:      "merch_2",
			AmountCents:     90000,
			Currency:        "GBP",
			BankAccountHash: "bank_xyz",
			DeviceID:        "dev_new",
			AccountAgeHours: 48,
			CreatedAt:       baseTime,
		}
		history := []PayoutRequest{
			{RequestID: "h1", AmountCents: 20000, DeviceID: "dev_old", CreatedAt: baseTime.Add(-2 * time.Hour)},
		}

		res, err := engine.Evaluate(context.Background(), req, history)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Decision != DecisionReview {
			t.Errorf("expected REVIEW, got %s (score: %d)", res.Decision, res.TotalScore)
		}
	})

	t.Run("DeclinesHighRiskPayout", func(t *testing.T) {
		req := PayoutRequest{
			RequestID:       "req_danger",
			MerchantID:      "merch_3",
			AmountCents:     120000,
			Currency:        "GBP",
			BankAccountHash: "bank_changed",
			DeviceID:        "dev_alien",
			AccountAgeHours: 4, // Tamper flag: 50
			CreatedAt:       baseTime,
		}
		history := []PayoutRequest{
			{RequestID: "h1", AmountCents: 15000, DeviceID: "dev_1", CreatedAt: baseTime.Add(-10 * time.Minute)},
			{RequestID: "h2", AmountCents: 25000, DeviceID: "dev_1", CreatedAt: baseTime.Add(-5 * time.Minute)},
			{RequestID: "h3", AmountCents: 70000, DeviceID: "dev_1", CreatedAt: baseTime.Add(-2 * time.Minute)}, // Velocity flag: 45
		}

		res, err := engine.Evaluate(context.Background(), req, history)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Decision != DecisionDecline {
			t.Errorf("expected DECLINE, got %s (score: %d)", res.Decision, res.TotalScore)
		}
		if res.TotalScore < 70 {
			t.Errorf("expected score >= 70, got %d", res.TotalScore)
		}
	})

	t.Run("ConcurrentEvaluations", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				req := PayoutRequest{
					RequestID:       "req_conc",
					MerchantID:      "merch_conc",
					AmountCents:     30000,
					Currency:        "GBP",
					BankAccountHash: "hash",
					DeviceID:        "dev",
					AccountAgeHours: 100,
					CreatedAt:       baseTime,
				}
				_, err := engine.Evaluate(context.Background(), req, nil)
				if err != nil {
					t.Errorf("concurrent evaluate error: %v", err)
				}
			}(i)
		}
		wg.Wait()
	})
}
