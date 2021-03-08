package events

import (
	"github.com/HOWZ1T/space_trader/models"
	"time"
)

type ShipOrder struct {
	name string
	time time.Time

	Type  string
	Order models.ShipOrder
}

func (e ShipOrder) New(args ...interface{}) Event {
	return &ShipOrder{
		name:  "SHIP_PURCHASED",
		time:  time.Now(),
		Type:  args[0].(string),
		Order: args[1].(models.ShipOrder),
	}
}

func (e *ShipOrder) Name() string {
	return e.name
}

func (e *ShipOrder) Time() time.Time {
	return e.time
}
