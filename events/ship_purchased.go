package events

import (
	"github.com/HOWZ1T/space_trader/models"
	"time"
)

type ShipPurchased struct {
	name string
	time time.Time

	Account models.Account
}

func (e ShipPurchased) New(args ...interface{}) Event {
	return &ShipPurchased{
		name:    "SHIP_PURCHASED",
		time:    time.Now(),
		Account: args[0].(models.Account),
	}
}

func (e *ShipPurchased) Name() string {
	return e.name
}

func (e *ShipPurchased) Time() time.Time {
	return e.time
}
