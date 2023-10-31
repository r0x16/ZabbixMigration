package module

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

type ZabbixServerModule struct {
	Bundle *drivers.ApplicationBundle
}

var _ domain.ApplicationModule = &ZabbixServerModule{}

// Setup ZabbixServer module routes
func (m *ZabbixServerModule) Setup() {
	fmt.Println("ZabbixServerModule")
	zserver := m.Bundle.Server.Group("/zbxsrv")

	zserver.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Zabbix Server")
	})
}
