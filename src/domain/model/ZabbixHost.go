package model

import "gorm.io/gorm"

type ZabbixHost struct {
	gorm.Model
	HostID      string `json:"hostid" gorm:"type:varchar(255);not null"`
	Host        string `json:"host" gorm:"type:varchar(255);not null"`
	ProxyHostID string `json:"proxy_hostid" gorm:"type:varchar(255); index"`
	Status      string `json:"status" gorm:"not null"`

	// Migration info
	MigrationID uint       `gorm:"not null"`
	Migration   *Migration `gorm:"foreignKey:MigrationID"`

	Disabled int `gorm:"not null"`

	// Template Info
	Templates []*ZabbixTemplate `json:"parentTemplates" gorm:"-"`
}
