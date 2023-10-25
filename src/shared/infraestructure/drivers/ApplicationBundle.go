package drivers

import (
	"github.com/labstack/echo/v4"
	"git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure/drivers/db"
)

type ApplicationBundle struct {
	Server   *echo.Echo
	Database *db.GormPostgresDatabaseProvider
}

type ActionCallback func(echo.Context, *ApplicationBundle) error

func (bundle *ApplicationBundle) ActionInjection(callback ActionCallback) echo.HandlerFunc {
	return func(c echo.Context) error {
		return callback(c, bundle)
	}
}
