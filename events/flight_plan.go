package events

import (
	"github.com/HOWZ1T/space_trader/models"
	"time"
)

const T_CREATED = "CREATED"
const T_ENDED = "ENDED"

type FlightPlan struct {
	name string
	time time.Time

	Type string
	Plan models.FlightPlan
}

func (e FlightPlan) New(args ...interface{}) Event {
	return &FlightPlan{
		name: "FLIGHT_PLAN",
		time: time.Now(),
		Type: args[0].(string),
		Plan: args[1].(models.FlightPlan),
	}
}

func (e *FlightPlan) Name() string {
	return e.name
}

func (e *FlightPlan) Time() time.Time {
	return e.time
}
