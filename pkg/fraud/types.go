package fraud

import "time"

type Decision string

const (
	DecisionApprove Decision = "APPROVE"
	DecisionReview  Decision = "REVIEW"
	DecisionDecline Decision = "DECLINE"
)

type PayoutRequest struct {
	RequestID        string    `json:"request_id"`
	MerchantID       string    `json:"merchant_id"`
	RecipientID      string    `json:"recipient_id"`
	AmountCents      int64     `json:"amount_cents"`
	Currency         string    `json:"currency"`
	BankAccountHash  string    `json:"bank_account_hash"`
	DeviceID         string    `json:"device_id"`
	AccountAgeHours  int       `json:"account_age_hours"`
	CreatedAt        time.Time `json:"created_at"`
}

type RiskSignal struct {
	RuleName string   `json:"rule_name"`
	Score    int      `json:"score"` // 0 to 100
	Reason   string   `json:"reason"`
	Flagged  bool     `json:"flagged"`
}

type EvaluationResult struct {
	RequestID   string       `json:"request_id"`
	Decision    Decision     `json:"decision"`
	TotalScore  int          `json:"total_score"`
	Signals     []RiskSignal `json:"signals"`
	EvaluatedAt time.Time    `json:"evaluated_at"`
}
