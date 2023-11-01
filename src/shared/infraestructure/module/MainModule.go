package module

import (
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

type MainModule struct {
	Bundle *drivers.ApplicationBundle
}

var _ domain.ApplicationModule = &MainModule{}

// Setups base main module routes
func (m *MainModule) Setup() {
	// This is a simple GET route that in Develpoment is used to test things
	m.Bundle.Server.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", echo.Map{
			"title": "Index title!",
		})
	})

	// This route checks health of the application
	m.Bundle.Server.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
}
