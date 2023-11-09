package infraestructure

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/app"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/db"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/framework"
)

type Main struct {
	mainService *app.MainService
}

func (m *Main) RunServices() {
	m.mainService = &app.MainService{}

	// Creates a new application bundle
	dbProvider := &db.GormPostgresDatabaseProvider{}
	app := &framework.EchoApplicationProvider{
		Bundle: &drivers.ApplicationBundle{
			Database:     dbProvider,
			ServerEvents: make(map[string]domain.ServerEventProvider),
		},
	}

	// Prepares the application bundle
	err := m.mainService.Run(
		app,
		dbProvider,
	)

	// Captures application error, must be logged.
	if err != nil {
		panic(err)
	}
}
