package action

import (
	"net/http"
	"strconv"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/zabbix-server/repository"
	"github.com/labstack/echo/v4"
)

func GetZabbixServerFromParam(c echo.Context, bundle *drivers.ApplicationBundle) (*model.ZabbixServer, *model.Error) {
	id, err := getZabbixServerParamId(c)

	if err != nil {
		return nil, err
	}

	return getZabbixServerFromId(id, bundle)

}

func getZabbixServerParamId(c echo.Context) (uint, *model.Error) {
	paramId := c.Param("zbxid")
	badRequestError := &model.Error{
		Code:    http.StatusBadRequest,
		Message: "Invalid migration id",
	}

	if paramId == "" {
		return 0, badRequestError
	}

	id, err := strconv.ParseUint(paramId, 10, 32)

	if err != nil {
		return 0, badRequestError
	}

	return uint(id), nil
}

func getZabbixServerFromId(id uint, bundle *drivers.ApplicationBundle) (*model.ZabbixServer, *model.Error) {
	serverRepository := repository.NewZabbixServerRepository(bundle.Database.Connection)
	server, err := serverRepository.GetByID(id)

	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error getting migration",
			Data:    err,
		}
	}

	return server, nil
}
