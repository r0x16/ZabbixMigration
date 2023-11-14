package zabbix

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
)

func ServerConnector(zabbix *model.ZabbixServer) domain.ZabbixConnectorProvider {
	switch zabbix.Version {
	case VERSION_64:
		return API64(zabbix.URL)
	case VERSION_40:
		return API40(zabbix.URL)
	}
	return nil
}
