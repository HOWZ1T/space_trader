package space_trader

import (
	"encoding/json"
	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
	"strings"
)

// Retrieves the available loans.
func (st *SpaceTrader) AvailableLoans() ([]models.Loan, error) {
	if v := st.cache.Fetch("available_loans"); v != nil {
		return v.([]models.Loan), nil
	}

	var raw map[string][]models.Loan
	err := st.doShaped("GET", loans, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return nil, err
	}

	st.cache.Store("available_loans", raw["loans"])
	return raw["loans"], nil
}

// Takes (purchases) a loan.
// Returns the updated Account info.
func (st *SpaceTrader) TakeLoan(loanType string) (models.Account, error) {
	// check loan type
	loanType = strings.Trim(loanType, "\r\n")
	switch strings.ToLower(loanType) {
	case "startup":
	case "enterprise":
		break

	default:
		return models.Account{}, errs.New("Invalid Option", "invalid loan type: "+loanType)
	}

	uri := users + st.username + "/loans"
	byts, err := json.Marshal(map[string]string{
		"type": strings.ToUpper(loanType),
	})
	if err != nil {
		return models.Account{}, err
	}

	var raw map[string]models.Account
	err = st.doShaped("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	st.eventManager.Emit(events.Loan{}.New(events.T_PURCHASED, raw["user"]))
	return raw["user"], nil
}

// Pays the specified loan.
// Returns the updated Account info.
func (st *SpaceTrader) PayLoan(loanID string) (models.Account, error) {
	uri := users + st.username + "/loans/" + loanID

	var raw map[string]models.Account
	err := st.doShaped("PUT", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	st.eventManager.Emit(events.Loan{}.New(events.T_PAID, raw["user"]))
	return raw["user"], nil
}
