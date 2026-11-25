# Financial Product Fraud Evaluator

High-throughput, real-time risk evaluation engine in Go for instant payout rails, merchant cash advances, and earned-wage transactions.

## Capabilities
- Concurrent rule evaluation pipeline with configurable risk weights.
- Rolling velocity and anomaly scoring across payout destinations.
- Bank account / device tampering detection.
- Sub-millisecond deterministic risk decisions (APPROVE, REVIEW, DECLINE).

## Usage
```bash
go test -v ./...
```
