package src

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/module"
)

func ProvideModules(bundle *drivers.ApplicationBundle) []domain.ApplicationModule {
	return []domain.ApplicationModule{
		&module.MainModule{Bundle: bundle},
	}
}
