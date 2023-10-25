package app

import "git.tnschile.com/sistemas/tnsgo/raidark/src/shared/domain"

type MainService struct {
}

/**
 * Boot the application and run the setup for all modules
 */
func (main *MainService) Run(app domain.ApplicationProvider, db domain.DatabaseProvider) error {
	err := db.Connect()
	if err != nil {
		return err
	}

	return main.applicationBootstraping(app)
}

/**
 * Initializes the application
 */
func (main *MainService) applicationBootstraping(app domain.ApplicationProvider) error {
	app.Boot()
	main.moduleSetup(app.ProvideModules())
	return app.Run()
}

/**
 * Setup all modules
 */
func (main *MainService) moduleSetup(modules []domain.ApplicationModule) {
	for _, module := range modules {
		module.Setup()
	}
}
