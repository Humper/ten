package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name       string `gorm:"not null"`
	Email      string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	Role       string
	AllowedIPs pq.StringArray `gorm:"column:allowed_ips;type:text[]"`
}
