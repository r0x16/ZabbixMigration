package model

import "gorm.io/gorm"

type ZabbixServer struct {
	gorm.Model
	URL      string        `json:"url" gorm:"unique;not null"`
	Username string        `json:"username" gorm:"not null"`
	Password string        `json:"password" gorm:"not null"`
	Version  ZabbixVersion `json:"version" gorm:"not null"`
}

type ZabbixVersion int
