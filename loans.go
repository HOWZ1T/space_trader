package space_trader

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
)

// Retrieves the available loans.
func (st *SpaceTrader) AvailableLoans() ([]models.Loan, error) {
	if v := st.cache.Fetch("available_loans"); v != nil {
		return v.([]models.Loan), nil
	}

	var raw map[string][]models.Loan
	err := st.doShaped("GET", loans, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
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
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
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
	err := st.doShaped("PUT", uri, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	st.eventManager.Emit(events.Loan{}.New(events.T_PAID, raw["user"]))
	return raw["user"], nil
}
