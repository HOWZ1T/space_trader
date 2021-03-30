package space_trader

import (
	"encoding/json"
	"fmt"

	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
)

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
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return models.FlightPlan{}, err
	}

	st.mu.Lock()
	st.flightPlans[raw["flightPlan"].ID] = raw["flightPlan"]
	st.mu.Unlock()
	st.eventManager.Emit(events.FlightPlan{}.New(events.T_CREATED, raw["flightPlan"]))
	return raw["flightPlan"], nil
}

// Retrieves a FlightPlan by the given ID.
func (st *SpaceTrader) GetFlightPlan(flightPlanID string) (models.FlightPlan, error) {
	uri := users + st.username + "/flight-plans/" + flightPlanID

	var raw map[string]models.FlightPlan
	err := st.doShaped("GET", uri, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return models.FlightPlan{}, err
	}

	return raw["flightPlan"], nil
}

// Retrieves all flight plans within a system
// Note you must have an active ship in that system
func (st *SpaceTrader) GetAllFlightPlansWithinSystem(symbol string) ([]models.CommonFlightPlan, error) {
	uri := systems + symbol + "/flight-plans/"

	var raw map[string][]models.CommonFlightPlan
	err := st.doShaped("GET", uri, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return []models.CommonFlightPlan{}, err
	}

	return raw["flightPlans"], nil
}
