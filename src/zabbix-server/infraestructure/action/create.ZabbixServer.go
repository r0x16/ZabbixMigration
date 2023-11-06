package action

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/zabbix-server/repository"
	"github.com/labstack/echo/v4"
)

func CreateZabbixServer(c echo.Context, bundle *drivers.ApplicationBundle) error {
	var dataError *model.Error
	var zabbixServer *model.ZabbixServer
	if c.Request().Method == http.MethodPost {
		zabbixServer, dataError = storeZabbixServer(c, bundle)
	}

	zabbixServers, listError := listZabbixServer(bundle)
	if listError != nil {
		c.Logger().Panic(dataError)
		return echo.NewHTTPError(listError.Code, listError.Message)
	}

	return c.Render(http.StatusOK, "zabbix-server/create", echo.Map{
		"title":            "Server administration",
		"error":            dataError,
		"zabbixServer":     zabbixServer,
		"zabbixServerList": zabbixServers,
	})
}

func listZabbixServer(bundle *drivers.ApplicationBundle) ([]*model.ZabbixServer, *model.Error) {

	zabbixServerRepository := repository.NewZabbixServerRepository(bundle.Database.Connection)
	zabbixServers, err := zabbixServerRepository.GetAll()
	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error listing zabbix servers",
			Data:    zabbixServers,
		}
	}

	return zabbixServers, nil
}

func storeZabbixServer(c echo.Context, bundle *drivers.ApplicationBundle) (*model.ZabbixServer, *model.Error) {

	zabbixServer, dataError := bindZabbixServer(c)
	if dataError != nil {
		return nil, dataError
	}

	if connectionError := testingZabbixConnection(zabbixServer); connectionError != nil {
		return nil, connectionError
	}

	if credentialsError := validateCredentials(zabbixServer); credentialsError != nil {
		return nil, credentialsError
	}

	fmt.Println(zabbixServer)

	zabbixServerRepository := repository.NewZabbixServerRepository(bundle.Database.Connection)
	if err := zabbixServerRepository.Store(zabbixServer); err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error storing zabbix server",
			Data:    zabbixServer,
		}
	}

	return zabbixServer, nil
}

func bindZabbixServer(c echo.Context) (*model.ZabbixServer, *model.Error) {

	var zabbixServer model.ZabbixServer
	if err := c.Bind(&zabbixServer); err != nil {
		return nil, &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Error in request data",
			Data:    zabbixServer,
		}
	}

	return &zabbixServer, nil
}

func testingZabbixConnection(zabbixServer *model.ZabbixServer) *model.Error {
	connector := zabbix.API(zabbixServer.URL)
	version, err := connector.GetVersion()

	if err != nil {
		return err
	}

	zabbixServer.Version = version

	return nil
}

func validateCredentials(zabbixServer *model.ZabbixServer) *model.Error {
	var connector domain.ZabbixConnectorProvider
	switch zabbixServer.Version {
	case zabbix.VERSION_40:
		connector = zabbix.API40(zabbixServer.URL)
	case zabbix.VERSION_60:
		connector = zabbix.API64(zabbixServer.URL)
	default:
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Zabbix version not supported",
			Data:    zabbixServer,
		}
	}

	if err := connector.Connect(zabbixServer.Username, zabbixServer.Password); err != nil {
		return err
	}

	return nil
}
