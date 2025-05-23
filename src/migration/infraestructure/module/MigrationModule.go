package module

import (
	"os"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action/runjob"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action/tplmap"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
)

type MigrationModule struct {
	Bundle *drivers.ApplicationBundle
}

func (m *MigrationModule) Setup() {
	os.MkdirAll("logs/migration", 0755)
	zserver := m.Bundle.Server.Group("/migration")

	zserver.GET("", m.Bundle.ActionInjection(action.CreateMigration)).Name = "MigrationCreate"
	zserver.POST("", m.Bundle.ActionInjection(action.CreateMigration)).Name = "MigrationStore"

	zserver.GET("/:id/proxy-map", m.Bundle.ActionInjection(action.SetupProxyMapping)).Name = "ProxyMapFlow"
	zserver.POST("/:id/proxy-map", m.Bundle.ActionInjection(action.SetupProxyMapping)).Name = "ProxyMapFlow_store"
	zserver.GET("/:id/proxy-map/import-events", m.Bundle.ActionInjection(action.ImportProxyStatusEvents)).Name = "ProxyMapFlow_importStatus"

	zserver.GET("/:id/template-map", m.Bundle.ActionInjection(tplmap.Setup)).Name = "TemplateMapFlow"
	zserver.POST("/:id/template-map", m.Bundle.ActionInjection(tplmap.Setup)).Name = "TemplateMapFlow_store"
	zserver.GET("/:id/template-map/import-events", m.Bundle.ActionInjection(tplmap.ImportStatus)).Name = "TemplateMapFlow_importStatus"

	zserver.GET("/:id/run", m.Bundle.ActionInjection(runjob.Run)).Name = "StartMigrationFlow"
	zserver.POST("/:id/run", m.Bundle.ActionInjection(runjob.Run)).Name = "StartMigrationFlow_Run"
	zserver.GET("/:id/run/status/:logLines", m.Bundle.ActionInjection(runjob.RunStatus)).Name = "StartMigrationFlow_RunStatus"
}
