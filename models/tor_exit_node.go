package models

import (
	"gorm.io/gorm"
)

type TorExitNode struct {
	gorm.Model
	IP          string `gorm:"unique;not null"`
	CountryName string `json:"country_name"`
	CountryCode string `json:"country_code"`
}
