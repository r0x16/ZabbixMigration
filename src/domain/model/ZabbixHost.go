package model

import "gorm.io/gorm"

type ZabbixHost struct {
	gorm.Model
	HostID  string `json:"hostid" gorm:"type:varchar(255);not null"`
	Host    string `json:"host" gorm:"type:varchar(255);not null"`
	ProxyID string `json:"proxy_hostid" gorm:"type:varchar(255)"`
	Status  int    `json:"status" gorm:"not null"`

	Disabled int `gorm:"not null"`
}
