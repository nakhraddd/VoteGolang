package models

import "gorm.io/gorm"

type Deputy struct {
	gorm.Model
	Name      string `json:"name"`
	Photo     string `json:"photo"`
	Education string `json:"education"`
	Age       int    `json:"age"`
	Party     string `json:"party"`
	Region    string `json:"region"`
}
