package model

import "gorm.io/gorm"

type ZabbixServer struct {
	gorm.Model
	Name     string        `json:"name" gorm:"unique;not null" form:"connectionName"`
	URL      string        `json:"url" gorm:"unique;not null" form:"apiUrl"`
	Username string        `json:"username" gorm:"not null" form:"username"`
	Password string        `json:"password" gorm:"not null" form:"password"`
	Version  ZabbixVersion `json:"version" gorm:"not null"`
}

type ZabbixVersion int

const (
	VERSION_UNKNOWN ZabbixVersion = 0
)
