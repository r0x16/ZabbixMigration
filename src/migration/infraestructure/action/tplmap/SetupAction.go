package tplmap

import (
	"fmt"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

type SetupAction struct {
	c      echo.Context
	bundle *drivers.ApplicationBundle
}

func Setup(c echo.Context, bundle *drivers.ApplicationBundle) error {
	setup := &SetupAction{c: c, bundle: bundle}

	migration, err := CheckTemplateImport(c, bundle)
	if err != nil {
		return echo.NewHTTPError(err.Code, err.Message)
	}

	fmt.Println(migration, setup)
	return nil
}
