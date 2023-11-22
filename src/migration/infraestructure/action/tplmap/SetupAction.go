package tplmap

import (
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

type SetupAction struct {
	c         echo.Context
	bundle    *drivers.ApplicationBundle
	migration *model.Migration
}

func Setup(c echo.Context, bundle *drivers.ApplicationBundle) error {
	setup := &SetupAction{c: c, bundle: bundle}

	migrationError := setup.SetupMigration()
	if migrationError != nil {
		return echo.NewHTTPError(migrationError.Code, migrationError.Message)
	}

	var storeError *model.Error
	if c.Request().Method == http.MethodPost {
		storeError = Store(setup.migration, c, bundle)
		if storeError == nil {
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("MigrationCreate", setup.migration.ID))
		}
	}

	var templateData *MappingBase
	var templateError *model.Error
	if setup.migration.IsTemplateImported {
		templateData, templateError = SetupBaseMapping(setup.bundle, setup.migration)
		if templateError != nil {
			return echo.NewHTTPError(templateError.Code, templateError.Message)
		}
	}

	importEventsUrl := c.Echo().Reverse("TemplateMapFlow_importStatus", setup.migration.ID)
	return c.Render(http.StatusOK, "migration/template-map", echo.Map{
		"title":           "Template mapping",
		"migration":       setup.migration,
		"importEventsUrl": importEventsUrl,
		"templateData":    templateData,
		"error":           storeError,
	})
}

func (s *SetupAction) SetupMigration() *model.Error {
	migration, migrationError := CheckTemplateImport(s.c, s.bundle)
	if migrationError != nil {
		return migrationError
	}
	s.migration = migration
	return nil
}
