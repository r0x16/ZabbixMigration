package action

import (
	"net/http"
	"strconv"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

func GetMigrationFromParam(c echo.Context, bundle *drivers.ApplicationBundle) (*model.Migration, *model.Error) {
	id, err := getMigrationParamId(c)

	if err != nil {
		return nil, err
	}

	return getMigrationFromId(id, bundle)
}

func getMigrationParamId(c echo.Context) (uint, *model.Error) {
	paramId := c.Param("id")
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

func getMigrationFromId(id uint, bundle *drivers.ApplicationBundle) (*model.Migration, *model.Error) {
	migrationRepository := repository.NewMigrationRepository(bundle.Database.Connection)
	migration, err := migrationRepository.GetById(id)

	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error getting migration",
			Data:    err,
		}
	}

	return migration, nil
}
