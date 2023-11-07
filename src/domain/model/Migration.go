package model

import (
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	gorm.Model
	Name string `json:"name" gorm:"type:varchar(255);not null" form:"migrationName"`

	SourceID      uint         `json:"sourceId" gorm:"not null" form:"sourceServer"`
	Source        ZabbixServer `json:"source" gorm:"foreignKey:SourceID"`
	DestinationID uint         `json:"destinationId" gorm:"not null" form:"destinationServer"`
	Destination   ZabbixServer `json:"destination" gorm:"foreignKey:DestinationID"`

	IsSuccess           bool      `gorm:"not null"`
	LastRunAt           time.Time `gorm:"not null"`
	IsProxyMapped       bool      `gorm:"not null"`
	HasTemplateBindings bool      `gorm:"not null"`
}
