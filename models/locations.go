package models

type PurchaseLocation struct {
	Location string   `json:"location"`
	Price    Currency `json:"price"`
}

type Location struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}
