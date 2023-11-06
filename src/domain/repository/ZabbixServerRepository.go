package repository

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ZabbixServerRepository interface {
	Store(zabbixServer *model.ZabbixServer) error
	GetAll() ([]*model.ZabbixServer, error)
}
