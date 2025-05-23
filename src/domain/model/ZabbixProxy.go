package model

import (
	"database/sql"

	"gorm.io/gorm"
)

type ZabbixProxy struct {
	gorm.Model
	Host         string        `json:"host" gorm:"type:varchar(255);not null"`
	ProxyID      string        `json:"proxyid" gorm:"type:varchar(255); index"`
	Status       string        `json:"status" gorm:"not null"`
	LastAccess   string        `json:"lastaccess" gorm:"not null"`
	ProxyAddress string        `json:"proxy_address" gorm:"type:varchar(255)"`
	Hosts        []interface{} `json:"hosts" gorm:"-"`
	HostCount    int           `json:"host_count" gorm:"not null"`

	// Passive Proxy interface
	InterfaceID sql.NullInt32         `json:"interfaceid"`
	Interface   *ZabbixProxyInterface `json:"interface" gorm:"foreignKey:InterfaceID"`

	// Migration in which this proxy is mapped
	MigrationID uint       `json:"migrationId" gorm:"not null"`
	Migration   *Migration `json:"migration" gorm:"foreignKey:MigrationID"`

	// Zabbix Server in which this proxy is mapped
	ZabbixServerID uint          `json:"zabbixServerId" gorm:"not null"`
	ZabbixServer   *ZabbixServer `json:"zabbixServer" gorm:"foreignKey:ZabbixServerID"`

	// Mapping
	SourceMapping      *ZabbixProxyMapping `json:"sourceMapping" gorm:"foreignKey:SourceProxyID"`
	DestinationMapping *ZabbixProxyMapping `json:"destinationMapping" gorm:"foreignKey:DestinationProxyID"`

	// Running status
	IsHostsRunning   bool `gorm:"not null;default:false"`
	IsHostSuccessful bool `gorm:"not null;default:false"`
	IsHostImporting  bool `gorm:"not null;default:false"`
	IsHostImported   bool `gorm:"not null;default:false"`
	IsHostDisabling  bool `gorm:"not null;default:false"`
	IsHostDisabled   bool `gorm:"not null;default:false"`
	IsRollingBack    bool `gorm:"not null;default:false"`
}

type ZabbixProxyInterface struct {
	gorm.Model
	Dns         string `json:"dns" gorm:"type:varchar(255);not null"`
	Ip          string `json:"ip" gorm:"type:varchar(255);not null"`
	Port        int    `json:"port" gorm:"not null"`
	Interfaceid string `json:"interfaceid" gorm:"type:varchar(255);not null"`
}

type ZabbixProxyMapping struct {
	gorm.Model
	SourceProxyID      uint `json:"sourceProxyId" gorm:"not null"`
	SourceProxy        *ZabbixProxy
	DestinationProxyID uint `json:"destinationProxyId" gorm:"not null"`
	DestinationProxy   *ZabbixProxy
}
