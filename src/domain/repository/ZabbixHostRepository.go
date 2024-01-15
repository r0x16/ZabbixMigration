package repository

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ZabbixHostRepository interface {
	FindByMigration(migration *model.Migration) ([]*model.ZabbixHost, error)
	FindByMigrationAndProxy(migration *model.Migration, proxy *model.ZabbixProxy) ([]*model.ZabbixHost, error)
	MultipleStore(hosts []*model.ZabbixHost) error
}
