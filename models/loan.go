package models

import "time"

// Models the Loan object.
type Loan struct {
	Type               string   `json:"type"`
	Amount             Currency `json:"amount"`
	CollateralRequired bool     `json:"collateralRequired"`
	Rate               Currency `json:"rate"`
	TermInDays         int      `json:"termInDays"`
}

// Models a loan that has been purchased.
type PurchasedLoan struct {
	Due             time.Time `json:"due"`
	ID              string    `json:"id"`
	RepaymentAmount Currency  `json:"repaymentAmount"`
	Status          string    `json:"status"`
	Type            string    `json:"type"`
}
