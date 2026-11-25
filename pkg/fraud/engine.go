package fraud

import (
	"context"
	"sync"
	"time"
)

type Engine struct {
	rules           []Rule
	declineThreshold int
	reviewThreshold  int
}

func NewEngine(rules []Rule, declineThreshold, reviewThreshold int) *Engine {
	return &Engine{
		rules:           rules,
		declineThreshold: declineThreshold,
		reviewThreshold:  reviewThreshold,
	}
}

// Evaluate runs all rules concurrently and aggregates risk signals into a decision.
func (e *Engine) Evaluate(ctx context.Context, req PayoutRequest, history []PayoutRequest) (EvaluationResult, error) {
	type ruleResult struct {
		signal RiskSignal
		err    error
	}

	resultsChan := make(chan ruleResult, len(e.rules))
	var wg sync.WaitGroup

	for _, rule := range e.rules {
		wg.Add(1)
		go func(r Rule) {
			defer wg.Done()
			sig, err := r.Evaluate(ctx, req, history)
			resultsChan <- ruleResult{signal: sig, err: err}
		}(rule)
	}

	wg.Wait()
	close(resultsChan)

	var signals []RiskSignal
	totalScore := 0

	for res := range resultsChan {
		if res.err != nil {
			return EvaluationResult{}, res.err
		}
		signals = append(signals, res.signal)
		totalScore += res.signal.Score
	}

	decision := DecisionApprove
	if totalScore >= e.declineThreshold {
		decision = DecisionDecline
	} else if totalScore >= e.reviewThreshold {
		decision = DecisionReview
	}

	return EvaluationResult{
		RequestID:   req.RequestID,
		Decision:    decision,
		TotalScore:  totalScore,
		Signals:     signals,
		EvaluatedAt: time.Now().UTC(),
	}, nil
}
