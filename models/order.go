package models

type Order struct {
	Good         string   `json:"good"`
	PricePerUnit Currency `json:"pricePerUnit"`
	Quantity     int      `json:"quantity"`
	TotalPrice   Currency `json:"total"`
}

type ShipOrder struct {
	Credits Currency `json:"credits"`
	Orders  []Order
	Ship    Ship `json:"ship"`
}
