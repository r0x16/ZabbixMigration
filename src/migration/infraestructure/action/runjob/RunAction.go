package runjob

import (
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

type RunAction struct {
	Context      echo.Context
	Bundle       *drivers.ApplicationBundle
	Migration    *model.Migration
	TemplateRepo *repository.ZabbixTemplateRepository
}

func Run(c echo.Context, bundle *drivers.ApplicationBundle) error {
	run := &RunAction{
		Context:      c,
		Bundle:       bundle,
		TemplateRepo: repository.NewZabbixTemplateRepository(bundle.Database.Connection),
	}

	migrationError := run.setupMigration()
	if migrationError != nil {
		return echo.NewHTTPError(migrationError.Code, migrationError.Message)
	}

	templateMigration := NewTemplateMigration(run)
	templateMigrationInfo, templateMigrationInfoError := templateMigration.GetMigrationInfo()
	if templateMigrationInfoError != nil {
		return echo.NewHTTPError(templateMigrationInfoError.Code, templateMigrationInfoError.Message)
	}

	return c.Render(http.StatusOK, "migration/run", echo.Map{
		"migration":    run.Migration,
		"templateInfo": templateMigrationInfo,
	})
}

func (s *RunAction) setupMigration() *model.Error {
	migration, migrationError := action.GetMigrationFromParam(s.Context, s.Bundle)
	if migrationError != nil {
		return migrationError
	}

	if !migration.IsProxyMapped {
		return &model.Error{
			Code:    http.StatusForbidden,
			Message: "Proxy mapping not configured",
		}
	}

	if !migration.HasTemplateBindings {
		return &model.Error{
			Code:    http.StatusForbidden,
			Message: "Template mapping not configured",
		}
	}

	s.Migration = migration
	return nil
}
