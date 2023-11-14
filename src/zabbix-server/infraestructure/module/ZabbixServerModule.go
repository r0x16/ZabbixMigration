package module

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/zabbix-server/infraestructure/action"
)

type ZabbixServerModule struct {
	Bundle *drivers.ApplicationBundle
}

var _ domain.ApplicationModule = &ZabbixServerModule{}

// Setup ZabbixServer module routes
func (m *ZabbixServerModule) Setup() {
	zserver := m.Bundle.Server.Group("/zbxsrv")

	zserver.GET("", m.Bundle.ActionInjection(action.CreateZabbixServer)).Name = "ZabbixServerCreate"
	zserver.POST("", m.Bundle.ActionInjection(action.CreateZabbixServer)).Name = "ZabbixServerStore"
	zserver.GET("/:zbxid", m.Bundle.ActionInjection(action.ShowZabbixServer)).Name = "ZabbixServerShow"
}
