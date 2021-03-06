package models

// models a system
type System struct {
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	Locations []Location `json:"locations"`
}
