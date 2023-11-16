package domain

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ProxyMappingBody struct {
	DefaultProxy       uint   `form:"defaultProxy"`
	SourceProxies      []uint `form:"sourceProxy"`
	DestinationProxies []uint `form:"destinationProxy"`

	// Imported proxies in database
	ImportedSourceProxies      []*model.ZabbixProxy
	ImportedDestinationProxies []*model.ZabbixProxy
}
