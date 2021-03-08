package space_trader

import (
	"github.com/HOWZ1T/space_trader/models"
	"math"
)

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

// Calculates the distance between two locations.
func (st *SpaceTrader) Distance(a models.Location, b models.Location) float64 {
	type Point struct {
		x float64
		y float64
	}

	pA := Point{
		x: float64(a.X),
		y: float64(a.Y),
	}

	pB := Point{
		x: float64(b.X),
		y: float64(b.Y),
	}

	return math.Sqrt(math.Pow(pB.x-pA.x, 2) + math.Pow(pB.y-pA.y, 2))
}
