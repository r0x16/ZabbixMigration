package src

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	main "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/module"
	zserver "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/zabbix-server/infraestructure/module"
)

func ProvideModules(bundle *drivers.ApplicationBundle) []domain.ApplicationModule {
	return []domain.ApplicationModule{
		&main.MainModule{Bundle: bundle},
		&zserver.ZabbixServerModule{Bundle: bundle},
	}
}
