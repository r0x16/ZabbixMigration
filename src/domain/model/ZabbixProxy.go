package model

import "gorm.io/gorm"

type ZabbixProxy struct {
	gorm.Model
	Host         string `json:"host" gorm:"type:varchar(255);not null"`
	Status       int    `json:"status" gorm:"not null"`
	LastAccess   int    `json:"lastaccess" gorm:"not null"`
	ProxyAddress string `json:"proxy_address" gorm:"type:varchar(255)"`
	HostCount    int    `json:"host_count" gorm:"not null"`

	// Passive Proxy interface
	InterfaceID uint                 `json:"interfaceid"`
	Interface   ZabbixProxyInterface `json:"interface" gorm:"foreignKey:InterfaceID"`

	// Migration in which this proxy is mapped
	MigrationID uint      `json:"migrationId" gorm:"not null"`
	Migration   Migration `json:"migration" gorm:"foreignKey:MigrationID"`
}

type ZabbixProxyInterface struct {
	gorm.Model
	Dns         string `json:"dns" gorm:"type:varchar(255);not null"`
	Ip          string `json:"ip" gorm:"type:varchar(255);not null"`
	Port        int    `json:"port" gorm:"not null"`
	Interfaceid string `json:"interfaceid" gorm:"type:varchar(255);not null"`
}
