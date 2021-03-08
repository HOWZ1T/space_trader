package models

import "time"

// Models the Loan object.
type Loan struct {
	Type               string    `json:"type"`
	Amount             Currency  `json:"amount"`
	CollateralRequired bool      `json:"collateralRequired"`
	Rate               Currency  `json:"rate"`
	TermInDays         int       `json:"termInDays"`
	Due                time.Time `json:"due"`
	ID                 string    `json:"id"`
	RepaymentAmount    Currency  `json:"repaymentAmount"`
	Status             string    `json:"status"`
}
