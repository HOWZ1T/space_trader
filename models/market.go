package models

// models a sellable good.
type Good struct {
	QuantityAvailable int      `json:"quantityAvailable"`
	VolumePerUnit     int      `json:"volumePerUnit"`
	PricePerUnit      Currency `json:"pricePerUnit"`
	Symbol            string   `json:"symbol"`
}

// models a market.
type Market struct {
	Goods []Good `json:"marketplace"`
}
