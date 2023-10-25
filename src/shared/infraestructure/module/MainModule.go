package module

import (
	"net/http"

	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/domain"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/repository"
	"github.com/labstack/echo/v4"
)

type MainModule struct {
	Bundle *drivers.ApplicationBundle
}

var _ domain.ApplicationModule = &MainModule{}

// Setups base main module routes
func (m *MainModule) Setup() {
	// This is a simple GET route that store a random string in a list
	// and shows the last 10 values
	// It also checks if the last value was created less than 5 seconds ago
	// If so, it waits until 5 seconds have passed
	m.Bundle.Server.GET("/", func(c echo.Context) error {
		randomRepository := repository.NewRandomRepositoryGorm(m.Bundle.Database.Connection)
		random := randomRepository.GenerateRandomString(50)

		return c.JSONPretty(http.StatusOK, random, "  ")
	})

	// This route checks health of the application
	m.Bundle.Server.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
