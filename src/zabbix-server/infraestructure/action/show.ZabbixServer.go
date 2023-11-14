package action

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
	"github.com/labstack/echo/v4"
)

func ShowZabbixServer(c echo.Context, bundle *drivers.ApplicationBundle) error {
	zabbixServer, dataError := GetZabbixServerFromParam(c, bundle)
	if dataError != nil {
		return echo.NewHTTPError(dataError.Code, dataError.Message)
	}

	connector := zabbix.ServerConnector(zabbixServer)
	connectionError := connector.Connect(zabbixServer.Username, zabbixServer.Password)
	if connectionError != nil {
		return echo.NewHTTPError(connectionError.Code, connectionError.Message)
	}

	zabbixServerApiInfo, err := getZabbixServerApiInfo(connector)
	if err != nil {
		return echo.NewHTTPError(err.Code, err.Message)
	}

	return c.Render(http.StatusOK, "zabbix-server/show", echo.Map{
		"title":      fmt.Sprintf("Zabbix Server %s", zabbixServer.Name),
		"server":     zabbixServer,
		"apiVersion": zabbixServerApiInfo,
	})
}

func getZabbixServerApiInfo(connector domain.ZabbixConnectorProvider) (any, *model.Error) {
	data, err := connector.Request(connector.UnauthorizedBody("apiinfo.version", model.ZabbixParams{}))

	if err != nil {
		return nil, err
	}

	return data.Result, nil
}
