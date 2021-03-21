package models

// Models the PurchaseLocation object.
type PurchaseLocation struct {
	Location string   `json:"location"`
	Price    Currency `json:"price"`
}

// Models the Location object.
type Location struct {
	Name            string       `json:"name"`
	Symbol          string       `json:"symbol"`
	Type            string       `json:"type"`
	X               int          `json:"x"`
	Y               int          `json:"y"`
	Anomaly         string       `json:"anomaly"`
	AnsibleProgress int          `json:"ansibleProgress"`
	Ships           []DockedShip `json:"ships"`
	DockedShips     int          `json:"dockedShips"`
}

// Models the DockedShip object returned from a Location
type DockedShip struct {
	ID       string `json:"shipId"`
	Username string `json:"username"`
	Type     string `json:"shipType"`
}

// Models a Location object that has a Market.
type MarketLocation struct {
	Market
	Location
}
