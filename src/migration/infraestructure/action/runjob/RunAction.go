package runjob

import (
	"fmt"
	"net/http"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"github.com/gosimple/slug"
	"github.com/labstack/echo/v4"
)

type RunAction struct {
	Context           echo.Context
	Bundle            *drivers.ApplicationBundle
	Migration         *model.Migration
	MigrationRepo     *repository.MigrationRepository
	Log               *events.LogController
	TemplateRepo      *repository.ZabbixTemplateRepository
	TemplateMigration *TemplateMigration
	ProxyRepo         *repository.ZabbixProxyRepository
	HostRepo          *repository.ZabbixHostRepository
	HostImport        *HostImport
	HostMigration     *HostMigration
	HostDisable       *HostDisable
}

var runMutex sync.Mutex

func Run(c echo.Context, bundle *drivers.ApplicationBundle) error {
	run := &RunAction{
		Context:       c,
		Bundle:        bundle,
		TemplateRepo:  repository.NewZabbixTemplateRepository(bundle.Database.Connection),
		MigrationRepo: repository.NewMigrationRepository(bundle.Database.Connection),
		ProxyRepo:     repository.NewZabbixProxyRepository(bundle.Database.Connection),
		HostRepo:      repository.NewZabbixHostRepository(bundle.Database.Connection),
	}

	setupError := run.setup()
	if setupError != nil {
		return echo.NewHTTPError(setupError.Code, setupError.Message)
	}

	templateMigrationInfo, templateMigrationInfoError := run.TemplateMigration.GetMigrationInfo()
	if templateMigrationInfoError != nil {
		return echo.NewHTTPError(templateMigrationInfoError.Code, templateMigrationInfoError.Message)
	}

	hostMigrationInfo, HostMigrationInfoError := run.HostMigration.GetMigrationInfo()
	if HostMigrationInfoError != nil {
		return echo.NewHTTPError(HostMigrationInfoError.Code, HostMigrationInfoError.Message)
	}

	if c.Request().Method == http.MethodPost {
		runError := run.runPost()
		if runError != nil {
			return echo.NewHTTPError(runError.Code, runError.Message)
		}
		return c.Redirect(http.StatusFound, c.Echo().Reverse("StartMigrationFlow", run.Migration.ID))
	}

	currentLogs, currentLogsError := run.Log.GetCurrentLog()
	if currentLogsError != nil {
		return echo.NewHTTPError(currentLogsError.Code, currentLogsError.Message)
	}

	runEventsUrl := c.Echo().Reverse("StartMigrationFlow_RunStatus", run.Migration.ID, len(currentLogs))
	return c.Render(http.StatusOK, "migration/run", echo.Map{
		"migration":    run.Migration,
		"templateInfo": templateMigrationInfo,
		"hostInfo":     hostMigrationInfo,
		"currentLogs":  currentLogs,
		"runEventsUrl": runEventsUrl,
	})
}

func (s *RunAction) runPost() *model.Error {
	runMutex.Lock()
	defer runMutex.Unlock()

	runType := s.Context.FormValue("type")
	switch runType {
	case "template":
		if !s.Migration.IsTemplateRunning && !s.Migration.IsTemplateSuccessful {
			return s.TemplateMigration.Run()
		}
	case "host":
		if !s.Migration.IsRunning && !s.Migration.IsSuccess {
			return s.HostMigration.Run()
		}
	case "host-import":
		if !s.Migration.IsRunning && !s.Migration.IsSuccess {
			return s.HostImport.Run()
		}
	case "host-src-disable":
		if !s.Migration.IsRunning && !s.Migration.IsSuccess {
			return s.HostDisable.Run()
		}
	default:
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Invalid run type",
		}
	}

	return nil
}

func (s *RunAction) setup() *model.Error {
	migrationError := s.setupMigration()
	if migrationError != nil {
		return migrationError
	}

	logError := s.setupLogs()
	if logError != nil {
		return logError
	}

	s.TemplateMigration = NewTemplateMigration(s)
	s.HostImport = NewHostImport(s)
	s.HostMigration = NewHostMigration(s)
	s.HostDisable = NewHostDisable(s)

	return nil
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

func (s *RunAction) setupLogs() *model.Error {
	logFile := slug.Make(fmt.Sprintf("events-%s-%d", s.Migration.Name, s.Migration.ID))
	path := fmt.Sprintf("%s/%s.log", "logs/migration", logFile)
	log, logError := events.NewLogController(path)
	if logError != nil {
		return logError
	}
	s.Log = log
	return nil
}
