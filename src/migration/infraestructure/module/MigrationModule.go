package module

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
)

type MigrationModule struct {
	Bundle *drivers.ApplicationBundle
}

func (m *MigrationModule) Setup() {
	zserver := m.Bundle.Server.Group("/migration")

	zserver.GET("", m.Bundle.ActionInjection(action.CreateMigration)).Name = "MigrationCreate"
	zserver.POST("", m.Bundle.ActionInjection(action.CreateMigration)).Name = "MigrationStore"
}
