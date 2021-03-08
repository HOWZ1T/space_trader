// Space Traders is an API game where players explore the stars in order to exploit for it's riches.
//
// More info available here: https://spacetraders.io/
//
// This project provides a golang wrapper for the api.
package space_trader

import (
	"encoding/json"
	"github.com/HOWZ1T/space_trader/cache"
	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

	// locations endpoint
	locations = game + "locations/"
)

// SpaceTrader is a struct representing the API wrapper and provides functionality for consuming the API.
type SpaceTrader struct {
	mu       sync.Mutex
	token    string
	username string

	client       http.Client
	cache        cache.Cache
	eventManager events.EventManager

	flightPlans map[string]models.FlightPlan
}

// Creates a new SpaceTrader instance.
func New(token string, username string) *SpaceTrader {
	st := SpaceTrader{
		mu:       sync.Mutex{},
		token:    token,
		username: username,
		client: http.Client{
			Transport: http.DefaultTransport,
			Timeout:   60 * time.Second,
		},
		cache:        cache.New(time.Minute * 10),
		eventManager: events.NewManager(),
		flightPlans:  make(map[string]models.FlightPlan),
	}
	go st.tick()
	return &st
}

func (st *SpaceTrader) tick() {
	for {
		// check flight plans
		if len(st.flightPlans) > 0 {
			now := time.Now()
			st.mu.Lock()
			for k, v := range st.flightPlans {
				if now.After(v.ArrivesAt) || now.Equal(v.ArrivesAt) {
					f, err := st.GetFlightPlan(v.ID)
					if err != nil {
						panic(err)
					}

					// remove flight plan from internal map
					delete(st.flightPlans, k)

					// emit event
					st.eventManager.Emit(events.FlightPlan{}.New("ENDED", f))
				}
			}
			st.mu.Unlock()
		}
		time.Sleep(1 * time.Second)
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

	if resp.StatusCode == 429 {
		return errs.New2("Rate Limit Exceeded", "Too Many Requests", 429)
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

func (st *SpaceTrader) doShapedRateLimited(method string, uri string, body string, headers map[string]string,
	urlParams map[string]string, shape interface{}, maxWait float32, wait float32, retries int) error {
	req, err := st.newRequest(method, uri, body, headers, urlParams)
	if err != nil {
		return err
	}

	err = st.doRequestShaped(req, &shape)
	if err != nil {
		if e, ok := err.(*errs.ApiError); ok {
			if e.Code == 429 {
				// wait
				if retries > 0 && wait <= maxWait {
					time.Sleep(time.Duration(wait) * time.Second)
					wait = wait * 2.5
					retries--
					return st.doShapedRateLimited(method, uri, body, headers, urlParams, shape, maxWait, wait, retries)
				}
			}
		}
		return err
	}

	return nil
}

func (st *SpaceTrader) doShaped(method string, uri string, body string, headers map[string]string,
	urlParams map[string]string, shape interface{}) error {
	return st.doShapedRateLimited(method, uri, body, headers, urlParams, shape, 40, 1, 4)
}

func (st *SpaceTrader) EventsChannel() chan events.Event {
	return st.eventManager.EventChannel
}

// Changes this instance of SpaceTrader to be the specified user.
func (st *SpaceTrader) SwitchUser(token string, username string) {
	event := events.UserSwitched{}.New(username, token)
	st.eventManager.Emit(event)
	st.token = token
	st.username = username
}

// Retrieves the status of the api.
func (st *SpaceTrader) ApiStatus() (string, error) {
	var stat map[string]string

	err := st.doShaped("GET", status, "", nil, nil, &stat)
	if err != nil {
		return "", err
	}

	return stat["status"], nil
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
	err := st.doShaped("GET", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	return raw["user"], nil
}

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

// Retrieves the user's loans.
func (st *SpaceTrader) MyLoans() ([]models.Loan, error) {
	if v := st.cache.Fetch("my_loans"); v != nil && !st.cache.IsOld("my_loans") {
		return v.([]models.Loan), nil
	}

	uri := users + st.username + "/loans"

	var raw map[string][]models.Loan
	err := st.doShaped("GET", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return nil, err
	}

	st.cache.Store("my_loans", raw["loans"])
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
	st.eventManager.Emit(events.Loan{}.New("PURCHASED", raw["user"]))
	return raw["user"], nil
}

// Retrieves the available ships.
//
// Ships can be filtered by specifying class.
//
// If the class is specified as "" (empty) then no filter is applied.
func (st *SpaceTrader) AvailableShips(class string) ([]models.Ship, error) {
	if v := st.cache.Fetch("available_ships"); v != nil {
		if class == "" {
			return v.([]models.Ship), nil
		}

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

		ships := make([]models.Ship, 0)
		for _, ship := range v.([]models.Ship) {
			if strings.ToUpper(ship.Class) == class {
				ships = append(ships, ship)
			}
		}

		return ships, nil
	}

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

	var raw map[string][]models.Ship
	err := st.doShaped("GET", ships, "", nil, urlParams, &raw)
	if err != nil {
		return nil, err
	}

	st.cache.Store("available_ships", raw["ships"])
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

	var raw map[string]models.Account
	err = st.doShaped("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	st.eventManager.Emit(events.ShipPurchased{}.New(raw["user"]))
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

	var shipOrder models.ShipOrder
	err = st.doShaped("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	}, &shipOrder)
	if err != nil {
		return models.ShipOrder{}, err
	}

	st.eventManager.Emit(events.ShipOrder{}.New("BUY", shipOrder))
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

	var shipOrder models.ShipOrder
	err = st.doShaped("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	}, &shipOrder)
	if err != nil {
		return models.ShipOrder{}, err
	}

	st.eventManager.Emit(events.ShipOrder{}.New("SELL", shipOrder))
	return shipOrder, nil
}

// Searches the system for the given type.
// Returns an array of locations.
func (st *SpaceTrader) SearchSystem(system string, type_ string) ([]models.Location, error) {
	uri := systems + system + "/locations"
	urlParams := make(map[string]string)
	urlParams["token"] = st.token
	urlParams["type"] = type_

	var raw map[string][]models.Location
	err := st.doShaped("GET", uri, "", nil, urlParams, &raw)
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

	var raw map[string]models.FlightPlan
	err = st.doShaped("POST", uri, string(byts), map[string]string{
		"Content-Type": "application/json",
	}, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.FlightPlan{}, err
	}

	st.mu.Lock()
	st.flightPlans[raw["flightPlan"].ID] = raw["flightPlan"]
	st.mu.Unlock()
	st.eventManager.Emit(events.FlightPlan{}.New("CREATED", raw["flightPlan"]))
	return raw["flightPlan"], nil
}

// Retrieves a FlightPlan by the given ID.
func (st *SpaceTrader) GetFlightPlan(flightPlanID string) (models.FlightPlan, error) {
	uri := users + st.username + "/flight-plans/" + flightPlanID

	var raw map[string]models.FlightPlan
	err := st.doShaped("GET", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.FlightPlan{}, err
	}

	return raw["flightPlan"], nil
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
	st.eventManager.Emit(events.Loan{}.New("PAID", raw["user"]))
	return raw["user"], nil
}

// Get info for the specified location.
// Location is specified by it's symbol.
func (st *SpaceTrader) GetLocation(symbol string) (models.Location, error) {
	uri := locations + symbol

	var raw map[string]models.Location
	err := st.doShaped("GET", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.Location{}, err
	}

	return raw["planet"], nil
}

// Get all locations in the specified system.
// System is specified by it's symbol.
func (st *SpaceTrader) GetLocationsInSystem(symbol string) ([]models.Location, error) {
	uri := systems + symbol + "/locations"

	var raw map[string][]models.Location
	err := st.doShaped("GET", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return nil, err
	}

	return raw["locations"], nil
}

// Gets the location's marketplace info.
// Location is specified by it's symbol.
func (st *SpaceTrader) GetMarket(symbol string) (models.Market, error) {
	uri := locations + symbol + "/marketplace"

	var raw map[string]models.MarketLocation
	err := st.doShaped("GET", uri, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return models.Market{}, err
	}

	return raw["planet"].Market, nil
}

// Gets all the systems info.
func (st *SpaceTrader) GetSystems() ([]models.System, error) {
	if v := st.cache.Fetch("systems"); v != nil && !st.cache.IsOld("systems") {
		return v.([]models.System), nil
	}

	var raw map[string][]models.System
	err := st.doShaped("GET", systems, "", nil, map[string]string{
		"token": st.token,
	}, &raw)
	if err != nil {
		return nil, err
	}

	st.cache.Store("systems", raw["systems"])
	return raw["systems"], nil
}
