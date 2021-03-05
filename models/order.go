package models

// Models the Order object.
type Order struct {
	Good         string   `json:"good"`
	PricePerUnit Currency `json:"pricePerUnit"`
	Quantity     int      `json:"quantity"`
	TotalPrice   Currency `json:"total"`
}

// Models an order that has been purchased or sold by a ship.
type ShipOrder struct {
	Credits Currency `json:"credits"`
	Orders  []Order  `json:"order"`
	Ship    Ship     `json:"ship"`
}
