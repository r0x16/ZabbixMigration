package domain

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type HostMigrationInfo struct {
	TotalCount int
	Proxies    []*model.ZabbixProxy
}
