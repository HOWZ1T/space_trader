package space_trader

import (
	"fmt"

	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
)

// Changes this instance of SpaceTrader to be the specified user.
func (st *SpaceTrader) SwitchUser(token string, username string) {
	event := events.UserSwitched{}.New(username, token)
	st.eventManager.Emit(event)
	st.token = token
	st.username = username
}

// Registers a new user and returns the new user's token.
func (st *SpaceTrader) RegisterUser(username string) (string, error) {
	uri := users + username + "/token"

	var raw map[string]interface{}
	err := st.doShaped("POST", uri, "", nil, nil, &raw)
	if err != nil {
		/*if strings.Contains(err.Error(), "invalid character") {
			return "", errs.New("username error", "username already taken")
		}*/
		return "", err
	}

	// return token
	if val, ok := raw["token"]; ok {
		st.eventManager.Emit(events.UserRegistered{}.New(username, raw["token"]))
		return val.(string), nil
	}

	return "", errs.New("unknown", "error occurred")
}

// Retrieves the user's account info.
func (st *SpaceTrader) Account() (models.Account, error) {
	if v := st.cache.Fetch("account"); v != nil {
		return v.(models.Account), nil
	}

	uri := users + st.username
	var raw map[string]models.Account
	err := st.doShaped("GET", uri, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	return raw["user"], nil
}

// Retrieves the user's loans.
func (st *SpaceTrader) MyLoans() ([]models.Loan, error) {
	if v := st.cache.Fetch("my_loans"); v != nil && !st.cache.IsOld("my_loans") {
		return v.([]models.Loan), nil
	}

	uri := users + st.username + "/loans"

	var raw map[string][]models.Loan
	err := st.doShaped("GET", uri, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return nil, err
	}

	st.cache.Store("my_loans", raw["loans"])
	return raw["loans"], nil
}
