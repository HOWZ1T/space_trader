package models

import "time"

type FlightPlan struct {
	ArrivesAt              time.Time `json:"arrivesAt"`
	Destination            string    `json:"destination"`
	FuelConsumed           int       `json:"fuelConsumed"`
	FuelRemaining          int       `json:"fuelRemaining"`
	ID                     string    `json:"id"`
	ShipID                 string    `json:"ship"`
	TerminatedAt           time.Time `json:"terminatedAt"`
	TimeRemainingInSeconds int       `json:"timeRemainingInSeconds"`
}
