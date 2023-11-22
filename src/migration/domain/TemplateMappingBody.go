package domain

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type TemplateMappingBody struct {
	SourceTemplates      []uint `form:"sourceTemplate"`
	DestinationTemplates []uint `form:"destinationTemplate"`

	// Imported templates in database
	ImportedSourceTemplates      []*model.ZabbixTemplate
	ImportedDestinationTemplates []*model.ZabbixTemplate
}
