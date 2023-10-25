package framework

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"git.tnschile.com/sistemas/tnsgo/raidark/src"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/domain"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers"
)

type EchoApplicationProvider struct {
	Bundle *drivers.ApplicationBundle
}

var _ domain.ApplicationProvider = &EchoApplicationProvider{}

// Creates a new Echo server to serve http requests and response
func (app *EchoApplicationProvider) Boot() {
	server := echo.New()
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	app.Bundle.Server = server
}

// Provides the list of Echo modules to bootstrap all the routes
func (app *EchoApplicationProvider) ProvideModules() []domain.ApplicationModule {
	return src.ProvideModules(app.Bundle)
}

// Runs the HTTP server in the especified port and listens to errors
func (app *EchoApplicationProvider) Run() error {
	err := app.Bundle.Server.Start(":" + os.Getenv("CRODONT_PORT"))
	// Start server
	app.Bundle.Server.Logger.Fatal(
		err,
	)

	return err
}
