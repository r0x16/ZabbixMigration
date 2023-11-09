package drivers

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/db"
	"github.com/labstack/echo/v4"
)

type ApplicationBundle struct {
	Server       *echo.Echo
	Database     *db.GormPostgresDatabaseProvider
	ServerEvents map[string]domain.ServerEventProvider
}

type ActionCallback func(echo.Context, *ApplicationBundle) error

func (bundle *ApplicationBundle) ActionInjection(callback ActionCallback) echo.HandlerFunc {
	return func(c echo.Context) error {
		return callback(c, bundle)
	}
}
