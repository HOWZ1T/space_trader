package space_trader

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/HOWZ1T/space_trader/events"
	"github.com/HOWZ1T/space_trader/models"
)

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
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &shipOrder)
	if err != nil {
		return models.ShipOrder{}, err
	}

	st.eventManager.Emit(events.ShipOrder{}.New(events.T_BUY, shipOrder))
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
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", st.token),
	}, nil, &shipOrder)
	if err != nil {
		return models.ShipOrder{}, err
	}

	st.eventManager.Emit(events.ShipOrder{}.New(events.T_SELL, shipOrder))
	return shipOrder, nil
}
