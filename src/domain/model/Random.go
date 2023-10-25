package model

import "gorm.io/gorm"

type Random struct {
	gorm.Model
	Value string `gorm:"unique;not null"`
}