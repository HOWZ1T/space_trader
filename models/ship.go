package models

// Models the Ship object.
type Ship struct {
	Class             string             `json:"class"`
	DockingEfficiency Modifier           `json:"dockingEfficiency"`
	FuelEfficiency    Modifier           `json:"fuelEfficiency"`
	Maintenance       Modifier           `json:"maintenance"`
	Manufacturer      string             `json:"manufacturer"`
	MaxCargo          int                `json:"maxCargo"`
	Plating           int                `json:"plating"`
	Speed             Modifier           `json:"speed"`
	Type              string             `json:"type"`
	Weapons           int                `json:"weapons"`
	PurchaseLocation  []PurchaseLocation `json:"purchaseLocations"`
	Cargo             []Cargo            `json:"cargo"`
	ID                string             `json:"id"`
	Location          string             `json:"location"`
	SpaceAvailable    int                `json:"spaceAvailable"`
}
