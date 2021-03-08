package models

// models a sellable good.
type Good struct {
	Available     int      `json:"available"`
	VolumePerUnit int      `json:"volumePerUnit"`
	PricePerUnit  Currency `json:"pricePerUnit"`
	Symbol        string   `json:"symbol"`
}

// models a market.
type Market struct {
	Goods []Good `json:"marketplace"`
}
