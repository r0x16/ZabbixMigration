package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	gorm.Model
	Name string `json:"name" gorm:"type:varchar(255);not null" form:"migrationName"`

	SourceID       uint          `json:"sourceId" gorm:"not null" form:"sourceServer"`
	Source         ZabbixServer  `json:"source" gorm:"foreignKey:SourceID"`
	DestinationID  uint          `json:"destinationId" gorm:"not null" form:"destinationServer"`
	Destination    ZabbixServer  `json:"destination" gorm:"foreignKey:DestinationID"`
	DefaultProxyID sql.NullInt32 `json:"defaultProxyId" form:"defaultProxy"`
	DefaultProxy   *ZabbixProxy  `json:"defaultProxy" gorm:"foreignKey:DefaultProxyID"`

	IsRunning            bool      `gorm:"not null;default:false"`
	IsTemplateSuccessful bool      `json:"isTemplateSuccessful" gorm:"not null;default:false"`
	IsSuccess            bool      `gorm:"not null"`
	LastRunAt            time.Time `gorm:"not null"`
	IsProxyImported      bool      `gorm:"not null;default:false"`
	IsProxyMapped        bool      `gorm:"not null"`
	IsTemplateImported   bool      `gorm:"not null;default:false"`
	HasTemplateBindings  bool      `gorm:"not null"`
}
