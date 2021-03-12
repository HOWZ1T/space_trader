package models

import "time"

// Models the FlightPlan object.
type FlightPlan struct {
	ArrivesAt              time.Time `json:"arrivesAt"`
	Destination            string    `json:"destination"`
	FuelConsumed           int       `json:"fuelConsumed"`
	FuelRemaining          int       `json:"fuelRemaining"`
	ID                     string    `json:"id"`
	ShipID                 string    `json:"ship"`
	TerminatedAt           time.Time `json:"terminatedAt"`
	TimeRemainingInSeconds int       `json:"timeRemainingInSeconds"`
	Departure              string    `json:"departure"`
	Distance               int       `json:"distance"`
}
