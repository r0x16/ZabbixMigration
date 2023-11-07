package action

import (
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	zbxsrv "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/zabbix-server/repository"
	"github.com/labstack/echo/v4"
)

func CreateMigration(c echo.Context, bundle *drivers.ApplicationBundle) error {

	var dataError *model.Error
	var migration *model.Migration
	if c.Request().Method == http.MethodPost {
		migration, dataError = storeMigration(c, bundle)
	}

	servers, listError := listZabbixServer(bundle)
	if listError != nil {
		c.Logger().Panic(listError)
		return echo.NewHTTPError(listError.Code, listError.Message)
	}

	migrations, listError := listMigrations(bundle)
	if listError != nil {
		c.Logger().Panic(listError)
		return echo.NewHTTPError(listError.Code, listError.Message)
	}

	return c.Render(http.StatusOK, "migration/create", echo.Map{
		"title":         "Server migration",
		"servers":       servers,
		"error":         dataError,
		"migration":     migration,
		"migrationList": migrations,
	})
}

func listMigrations(bundle *drivers.ApplicationBundle) ([]*model.Migration, *model.Error) {

	migrationRepository := repository.NewMigrationRepository(bundle.Database.Connection)
	migrations, err := migrationRepository.GetAll()
	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error listing migrations",
			Data:    migrations,
		}
	}

	return migrations, nil
}

func listZabbixServer(bundle *drivers.ApplicationBundle) ([]*model.ZabbixServer, *model.Error) {

	zabbixServerRepository := zbxsrv.NewZabbixServerRepository(bundle.Database.Connection)
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

func storeMigration(c echo.Context, bundle *drivers.ApplicationBundle) (*model.Migration, *model.Error) {

	migration, dataError := bindMigration(c)
	if dataError != nil {
		return nil, dataError
	}

	if validationError := validateServers(migration, bundle); validationError != nil {
		return migration, validationError
	}

	migrationRepository := repository.NewMigrationRepository(bundle.Database.Connection)
	if err := migrationRepository.Store(migration); err != nil {
		return migration, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error storing migration",
			Data:    err,
		}
	}

	return migration, nil
}

func bindMigration(c echo.Context) (*model.Migration, *model.Error) {

	var migration model.Migration
	if err := c.Bind(&migration); err != nil {
		return nil, &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Error binding migration",
			Data:    err,
		}
	}

	return &migration, nil
}

func validateServers(migration *model.Migration, bundle *drivers.ApplicationBundle) *model.Error {

	if sourceError := validateServer(migration.SourceID, bundle); sourceError != nil {
		sourceError.Message = "Invalid source server selected"
		return sourceError
	}

	if destinationError := validateServer(migration.DestinationID, bundle); destinationError != nil {
		destinationError.Message = "Invalid destination server selected"
		return destinationError
	}

	if migration.SourceID == migration.DestinationID {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Source and destination servers must be different",
			Data:    migration,
		}
	}

	return nil
}

func validateServer(serverId uint, bundle *drivers.ApplicationBundle) *model.Error {

	repo := zbxsrv.NewZabbixServerRepository(bundle.Database.Connection)
	server, err := repo.GetByID(serverId)
	if err != nil {
		return &model.Error{
			Code: http.StatusInternalServerError,
			Data: server,
		}
	}

	return nil
}
