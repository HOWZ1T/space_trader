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

	disallowUnknownFields bool
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
		cache:                 cache.New(time.Minute * 10),
		eventManager:          events.NewManager(),
		flightPlans:           make(map[string]models.FlightPlan),
		disallowUnknownFields: false,
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
					st.eventManager.Emit(events.FlightPlan{}.New(events.T_ENDED, f))
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

	decoder := json.NewDecoder(strings.NewReader(string(body)))
	if st.disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	err = decoder.Decode(shape)

	if err != nil {
		return err
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

// Retrieves the status of the api.
func (st *SpaceTrader) ApiStatus() (string, error) {
	var stat map[string]string

	err := st.doShaped("GET", status, "", nil, nil, &stat)
	if err != nil {
		return "", err
	}

	return stat["status"], nil
}
