package models

// Models the Account object.
type Account struct {
	Credits  int    `json:"credits"`
	Loans    []Loan `json:"loans"`
	Ships    []Ship `json:"ships"`
	Username string `json:"username"`
}
