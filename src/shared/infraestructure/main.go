package infraestructure

import (
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/app"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers/db"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers/framework"
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
			Database: dbProvider,
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
