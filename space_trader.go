// Space Traders is an API game where players explore the stars in order to exploit for it's riches.
//
// More info available here: https://spacetraders.io/
//
// This project provides a golang wrapper for the api.
package space_trader

import (
	"encoding/json"
	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/models"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// These constants represents the various api endpoints.
const (
	// base endpoint
	base = "https://api.spacetraders.io/"

	// game endpoint
	game = base + "game/"

	// status endpoint
	status = game + "status"

	// users endpoint
	users = base + "users/"

	// loans endpoint
	loans = game + "loans"

	// ships endpoint
	ships = game + "ships"

	// systems endpoint
	systems = game + "systems/"
)

// SpaceTrader is a struct representing the API wrapper and provides functionality for consuming the API.
type SpaceTrader struct {
	token    string
	username string

	client http.Client
}

// Creates a new SpaceTrader instance.
func New(token string, username string) SpaceTrader {
	return SpaceTrader{
		token:    token,
		username: username,
		client: http.Client{
			Transport: http.DefaultTransport,
			Timeout:   60 * time.Second,
		},
	}
}

// Internal method for making new requests.
func (st *SpaceTrader) newRequest(method string, uri string, body string, headers map[string]string,
	urlParams map[string]string) (*http.Request, error) {
	bodyReader := strings.NewReader(body)

	req, err := http.NewRequest(method, uri, bodyReader)
	if err != nil {
		return req, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	urlQuery := req.URL.Query()
	for k, v := range urlParams {
		urlQuery.Add(k, v)
	}

	req.URL.RawQuery = urlQuery.Encode()

	return req, nil
}

// Internal method for `do`-ing a request and attempting to unmarshal the response into the given shape.
func (st *SpaceTrader) doRequestShaped(req *http.Request, shape interface{}) error {
	resp, err := st.client.Do(req)
	if err != nil {
		return err
	}

	// safely defer close
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, shape)
	if err != nil {
		return err
	}

	// try check for json error returned
	var e errs.ApiError
	err1 := json.Unmarshal(body, &e)
	if err1 == nil && len(e.Error()) > 0 {
		return &e
	} else {
		var raw map[string]map[string]interface{}
		err1 = json.Unmarshal(body, &raw)
		if err1 == nil {
			if v, ok := raw["error"]; ok {
				e := errs.ApiError{
					Err:  "error",
					Msg:  v["message"].(string),
					Code: int(v["code"].(float64)),
				}
				return &e
			}
		}
	}

	return nil
}

// Changes this instance of SpaceTrader to be the specified user.
func (st *SpaceTrader) SwitchUser(token string, username string) {
	st.token = token
	st.username = username
}

// Retrieves the status of the api.
func (st *SpaceTrader) ApiStatus() (string, error) {
	var stat map[string]string

	req, err := st.newRequest("GET", status, "", nil, nil)
	if err != nil {
		return "", err
	}

	err = st.doRequestShaped(req, &stat)
	if err != nil {
		return "", err
	}

	return stat["status"], nil
}

// Registers a new user and returns the new user's token.
func (st *SpaceTrader) RegisterUser(username string) (string, error) {
	uri := users + username + "/token"
	req, err := st.newRequest("POST", uri, "", nil, nil)
	if err != nil {
		return "", err
	}

	var raw map[string]interface{}
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		if strings.Contains(err.Error(), "invalid character") {
			return "", errs.New("username error", "username already taken")
		}
		return "", err
	}

	// return token
	if val, ok := raw["token"]; ok {
		return val.(string), nil
	}

	return "", errs.New("unknown", "error occurred")
}

// Retrieves the user's account info.
func (st *SpaceTrader) Account() (models.Account, error) {
	uri := users + st.username
	req, err := st.newRequest("GET", uri, "", nil, map[string]string{
		"token": st.token,
	})

	if err != nil {
		return models.Account{}, err
	}

	var raw map[string]models.Account
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return models.Account{}, err
	}

	return raw["user"], nil
}

// Retrieves the available loans.
func (st *SpaceTrader) AvailableLoans() ([]models.Loan, error) {
	req, err := st.newRequest("GET", loans, "", nil, map[string]string{
		"token": st.token,
	})

	if err != nil {
		return nil, err
	}

	var raw map[string][]models.Loan
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return nil, err
	}

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

	req, err := st.newRequest("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	})

	if err != nil {
		return models.Account{}, err
	}

	var raw map[string]models.Account
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return models.Account{}, err
	}

	return raw["user"], nil
}

// Retrieves the available ships.
//
// Ships can be filtered by specifying class.
//
// If the class is specified as "" (empty) then no filter is applied.
func (st *SpaceTrader) AvailableShips(class string) ([]models.Ship, error) {
	urlParams := make(map[string]string)
	urlParams["token"] = st.token

	if class != "" {
		class = strings.Trim(class, "\r\n")
		class = strings.ToUpper(class)
		switch class {
		case "MK-III":
		case "MK-II":
		case "MK-I":
			break

		default:
			return nil, errs.New("invalid argument", "invalid class: "+class)
		}
		urlParams["class"] = class
	}

	req, err := st.newRequest("GET", ships, "", nil, urlParams)

	if err != nil {
		return nil, err
	}

	var raw map[string][]models.Ship
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return nil, err
	}

	return raw["ships"], nil
}

// Buys the specified ship.
// Returns the updated Account info.
func (st *SpaceTrader) BuyShip(location string, shipType string) (models.Account, error) {
	uri := users + st.username + "/ships"
	byts, err := json.Marshal(map[string]string{
		"location": location,
		"type":     strings.ToUpper(shipType),
	})

	if err != nil {
		return models.Account{}, err
	}

	req, err := st.newRequest("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}, map[string]string{
		"token": st.token,
	})
	if err != nil {
		return models.Account{}, err
	}

	var raw map[string]models.Account
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return models.Account{}, err
	}

	return raw["user"], nil
}

// Buys the specified good at the specified quantity for the specified ship.
// Returns ShipOrder
func (st *SpaceTrader) BuyGood(shipID string, good string, quantity int) (models.ShipOrder, error) {
	uri := users + st.username + "/purchase-orders"
	byts, err := json.Marshal(map[string]string{
		"shipId":   shipID,
		"good":     good,
		"quantity": strconv.Itoa(quantity),
	})

	if err != nil {
		return models.ShipOrder{}, err
	}

	req, err := st.newRequest("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	})
	if err != nil {
		return models.ShipOrder{}, err
	}

	var shipOrder models.ShipOrder
	err = st.doRequestShaped(req, &shipOrder)
	if err != nil {
		return models.ShipOrder{}, err
	}

	return shipOrder, nil
}

// Sells the specified good at the specified quantity for the specified ship.
// Returns ShipOrder
func (st *SpaceTrader) SellGood(shipID string, good string, quantity int) (models.ShipOrder, error) {
	uri := users + st.username + "/sell-orders"
	byts, err := json.Marshal(map[string]string{
		"shipId":   shipID,
		"good":     good,
		"quantity": strconv.Itoa(quantity),
	})

	if err != nil {
		return models.ShipOrder{}, err
	}

	req, err := st.newRequest("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	})
	if err != nil {
		return models.ShipOrder{}, err
	}

	var shipOrder models.ShipOrder
	err = st.doRequestShaped(req, &shipOrder)
	if err != nil {
		return models.ShipOrder{}, err
	}

	return shipOrder, nil
}

// Searches the system for the given type.
// Returns an array of locations.
func (st *SpaceTrader) SearchSystem(system string, type_ string) ([]models.Location, error) {
	uri := systems + system + "/locations"
	urlParams := make(map[string]string)
	urlParams["token"] = st.token
	urlParams["type"] = type_

	req, err := st.newRequest("GET", uri, "", nil, urlParams)
	if err != nil {
		return nil, err
	}

	var raw map[string][]models.Location
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return nil, err
	}

	return raw["locations"], nil
}

// Creates a flight plan for the ship to the destination.
// Remember to calculate your estimated fuel usage!
//
// Returns FlightPlan
func (st *SpaceTrader) CreateFlightPlan(shipID string, destination string) (models.FlightPlan, error) {
	uri := users + st.username + "/flight-plans"
	byts, err := json.Marshal(map[string]string{
		"shipId":      shipID,
		"destination": destination,
	})

	if err != nil {
		return models.FlightPlan{}, err
	}

	req, err := st.newRequest("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	})
	if err != nil {
		return models.FlightPlan{}, err
	}

	var raw map[string]models.FlightPlan
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return models.FlightPlan{}, err
	}

	return raw["flightPlan"], nil
}

// Retrieves a FlightPlan by the given ID.
func (st *SpaceTrader) GetFlightPlan(flightPlanID string) (models.FlightPlan, error) {
	uri := users + st.username + "/flight-plans/" + flightPlanID

	req, err := st.newRequest("GET", uri, "", nil, map[string]string{
		"token": st.token,
	})
	if err != nil {
		return models.FlightPlan{}, err
	}

	var raw map[string]models.FlightPlan
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return models.FlightPlan{}, err
	}

	return raw["flightPlan"], nil
}

// Pays the specified loan.
// Returns the updated Account info.
func (st *SpaceTrader) PayLoan(loanID string) (models.Account, error) {
	uri := users + st.username + "/loans/" + loanID

	req, err := st.newRequest("PUT", uri, "", nil, map[string]string{
		"token": st.token,
	})

	if err != nil {
		return models.Account{}, err
	}

	var raw map[string]models.Account
	err = st.doRequestShaped(req, &raw)
	if err != nil {
		return models.Account{}, err
	}

	return raw["user"], nil
}
