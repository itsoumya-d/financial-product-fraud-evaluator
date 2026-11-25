package fraud

import (
	"context"
	"fmt"
	"time"
)

// Rule defines the interface for independent fraud evaluation rules.
type Rule interface {
	Name() string
	Evaluate(ctx context.Context, req PayoutRequest, history []PayoutRequest) (RiskSignal, error)
}

// VelocityRule detects rapid payout attempts within a rolling window.
type VelocityRule struct {
	WindowDuration time.Duration
	MaxCount       int
	MaxTotalCents  int64
}

func (r *VelocityRule) Name() string {
	return "velocity_check"
}

func (r *VelocityRule) Evaluate(ctx context.Context, req PayoutRequest, history []PayoutRequest) (RiskSignal, error) {
	cutoff := req.CreatedAt.Add(-r.WindowDuration)
	count := 0
	var totalCents int64

	for _, h := range history {
		if h.CreatedAt.After(cutoff) {
			count++
			totalCents += h.AmountCents
		}
	}

	if count >= r.MaxCount || totalCents >= r.MaxTotalCents {
		return RiskSignal{
			RuleName: r.Name(),
			Score:    45,
			Reason:   fmt.Sprintf("velocity exceeded: %d payouts (%d cents) in last %s", count, totalCents, r.WindowDuration),
			Flagged:  true,
		}, nil
	}

	return RiskSignal{
		RuleName: r.Name(),
		Score:    0,
		Reason:   "velocity within acceptable threshold",
		Flagged:  false,
	}, nil
}

// AccountTamperRule detects recently changed bank accounts or newly introduced devices.
type AccountTamperRule struct {
	MinCooldownHours int
}

func (r *AccountTamperRule) Name() string {
	return "account_tamper_check"
}

func (r *AccountTamperRule) Evaluate(ctx context.Context, req PayoutRequest, history []PayoutRequest) (RiskSignal, error) {
	if req.AccountAgeHours < r.MinCooldownHours {
		return RiskSignal{
			RuleName: r.Name(),
			Score:    50,
			Reason:   fmt.Sprintf("destination bank account changed %d hours ago (cooldown %d hrs)", req.AccountAgeHours, r.MinCooldownHours),
			Flagged:  true,
		}, nil
	}

	// Check if device is completely new compared to historical requests
	if len(history) > 0 {
		seenDevice := false
		for _, h := range history {
			if h.DeviceID == req.DeviceID {
				seenDevice = true
				break
			}
		}
		if !seenDevice {
			return RiskSignal{
				RuleName: r.Name(),
				Score:    25,
				Reason:   "unrecognized device fingerprint for merchant",
				Flagged:  true,
			}, nil
		}
	}

	return RiskSignal{
		RuleName: r.Name(),
		Score:    0,
		Reason:   "account and device verified",
		Flagged:  false,
	}, nil
}

// AmountSpikeRule detects sudden payout amounts exceeding historic baselines.
type AmountSpikeRule struct {
	MaxMultiplier float64
}

func (r *AmountSpikeRule) Name() string {
	return "amount_spike_check"
}

func (r *AmountSpikeRule) Evaluate(ctx context.Context, req PayoutRequest, history []PayoutRequest) (RiskSignal, error) {
	if len(history) == 0 {
		return RiskSignal{
			RuleName: r.Name(),
			Score:    10,
			Reason:   "initial payout without historical baseline",
			Flagged:  false,
		}, nil
	}

	var sum int64
	for _, h := range history {
		sum += h.AmountCents
	}
	avg := float64(sum) / float64(len(history))
	if float64(req.AmountCents) > avg*r.MaxMultiplier {
		return RiskSignal{
			RuleName: r.Name(),
			Score:    35,
			Reason:   fmt.Sprintf("payout amount %d cents exceeds %.1fx average baseline (avg: %.0f cents)", req.AmountCents, r.MaxMultiplier, avg),
			Flagged:  true,
		}, nil
	}

	return RiskSignal{
		RuleName: r.Name(),
		Score:    0,
		Reason:   "payout amount aligns with historical volume",
		Flagged:  false,
	}, nil
}
