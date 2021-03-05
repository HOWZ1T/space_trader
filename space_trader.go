package space_trader

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"space_trader/errs"
	"space_trader/models"
	"strconv"
	"strings"
	"time"
)

const base = "https://api.spacetraders.io/"
const game = base + "game/"
const status = game + "status"
const users = base + "users/"
const loans = game + "loans"
const ships = game + "ships"
const systems = game + "systems/"

type SpaceTrader struct {
	token    string
	username string

	client http.Client
}

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

	var flightPlan models.FlightPlan
	err = st.doRequestShaped(req, &flightPlan)
	if err != nil {
		return models.FlightPlan{}, err
	}

	return flightPlan, nil
}

func (st *SpaceTrader) GetFlightPlan(flightPlanID string) (models.FlightPlan, error) {
	uri := users + st.username + "/flight-plans/" + flightPlanID

	req, err := st.newRequest("GET", uri, "", nil, map[string]string{
		"token": st.token,
	})
	if err != nil {
		return models.FlightPlan{}, err
	}

	var flightPlan models.FlightPlan
	err = st.doRequestShaped(req, &flightPlan)
	if err != nil {
		return models.FlightPlan{}, err
	}

	return flightPlan, nil
}

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
