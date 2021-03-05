package models

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
}

type OwnedShip struct {
	Cargo          []Cargo  `json:"cargo"`
	Class          string   `json:"class"`
	ID             string   `json:"id"`
	Location       string   `json:"location"`
	Manufacturer   string   `json:"manufacturer"`
	MaxCargo       int      `json:"maxCargo"`
	Plating        int      `json:"plating"`
	SpaceAvailable int      `json:"spaceAvailable"`
	Speed          Modifier `json:"speed"`
	Type           string   `json:"type"`
	Weapons        int      `json:"weapons"`
}
