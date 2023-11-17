package model

import "gorm.io/gorm"

type ZabbixTemplate struct {
	gorm.Model
	Templateid  string `json:"templateid" gorm:"type:varchar(255);not null"`
	Name        string `json:"name" gorm:"not null"`
	Host        string `json:"host" gorm:"not null"`
	Description string `json:"description" gorm:"not null"`

	// Migration in which this proxy is mapped
	MigrationID uint       `json:"migrationId" gorm:"not null"`
	Migration   *Migration `json:"migration" gorm:"foreignKey:MigrationID"`

	// Zabbix Server in which this proxy is mapped
	ZabbixServerID uint          `json:"zabbixServerId" gorm:"not null"`
	ZabbixServer   *ZabbixServer `json:"zabbixServer" gorm:"foreignKey:ZabbixServerID"`
}

type ZabbixTemplateMapping struct {
	gorm.Model
	SourceTemplateID      uint            `json:"sourceTemplateId" gorm:"not null"`
	SourceTemplate        *ZabbixTemplate `gorm:"foreignKey:SourceTemplateID"`
	DestinationTemplateID uint            `json:"destinationTemplateId" gorm:"not null"`
	DestinationTemplate   *ZabbixTemplate `gorm:"foreignKey:DestinationTemplateID"`
}
