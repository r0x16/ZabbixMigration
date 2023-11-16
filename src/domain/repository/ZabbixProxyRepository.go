package repository

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ZabbixProxyRepository interface {
	Store(zabbixProxy *model.ZabbixProxy) error
	GetAll() ([]*model.ZabbixProxy, error)
	GetByMigrationAndServer(migrationId uint, serverId uint) ([]*model.ZabbixProxy, error)
	MultipleStore(zabbixProxies []*model.ZabbixProxy) error
}
