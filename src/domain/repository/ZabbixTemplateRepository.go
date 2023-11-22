package repository

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ZabbixTemplateRepository interface {
	Store(zabbixTemplate *model.ZabbixTemplate) error
	GetAll() ([]*model.ZabbixTemplate, error)
	GetByMigrationAndServer(migrationId uint, serverId uint) ([]*model.ZabbixTemplate, error)
	MultipleStore(zabbixTemplates []*model.ZabbixTemplate) error
	StoreMapping(mapping *model.ZabbixTemplateMapping) error
}
