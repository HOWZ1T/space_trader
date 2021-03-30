package space_trader

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/HOWZ1T/space_trader/errs"
	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
)

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
	err := st.doShaped("GET", ships, "", map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, urlParams, &raw)
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
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &raw)
	if err != nil {
		return models.Account{}, err
	}

	st.cache.Store("account", raw["user"])
	st.eventManager.Emit(events.ShipPurchased{}.New(raw["user"]))
	return raw["user"], nil
}
