package runjob

import (
	"fmt"
	"net/http"
	"strconv"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/labstack/echo/v4"
)

type RunStatusAction struct {
	c         echo.Context
	bundle    *drivers.ApplicationBundle
	migration *model.Migration
	eventId   string
	clientId  string
	client    events.EventClientEcho
	Log       *events.LogController
}

func RunStatus(c echo.Context, bundle *drivers.ApplicationBundle) error {
	runStatus := &RunStatusAction{
		c:        c,
		bundle:   bundle,
		clientId: uuid.New().String(),
	}

	setupError := runStatus.Setup()
	if setupError != nil {
		return echo.NewHTTPError(setupError.Code, setupError.Message)
	}

	if !runStatus.migration.IsRunning {
		runStatus.client.Close()
		return c.NoContent(http.StatusOK)
	}

	runStatus.RunClient()
	runStatus.StopClient()
	return c.NoContent(http.StatusOK)
}

func (s *RunStatusAction) RunClient() {
	logLines, err := strconv.Atoi(s.c.Param("logLines"))
	if err != nil {
		logLines = 0
	}

	logDifference, _ := s.Log.GetLogFromLine(logLines)

	s.bundle.ServerEvents[s.eventId].Subscribe(s.client)
	s.client.SendMessage(&domain.EventMessage{
		Event: "log",
		Data:  logDifference,
	})
	s.client.Online()
}

func (s *RunStatusAction) StopClient() {
	eventHandler, active := s.bundle.ServerEvents[s.eventId]
	if active {
		eventHandler.Unsubscribe(s.client)
	}
}

func (s *RunStatusAction) Setup() *model.Error {
	setupError := s.SetupMigration()
	if setupError != nil {
		return setupError
	}

	setupClientError := s.SetupClient()
	if setupClientError != nil {
		return setupClientError
	}

	setupLogError := s.SetupLogs()
	if setupLogError != nil {
		return setupLogError
	}

	return nil
}

func (s *RunStatusAction) SetupMigration() *model.Error {
	runMutex.Lock()
	defer runMutex.Unlock()

	migration, importError := action.GetMigrationFromParam(s.c, s.bundle)
	if importError != nil {
		return importError
	}
	s.migration = migration
	s.eventId = fmt.Sprintf("migration-run-%d", migration.ID)
	return nil
}

func (s *RunStatusAction) SetupClient() *model.Error {
	client := events.NewEventClientEcho(s.clientId, s.c)
	clientError := client.Setup()
	if clientError != nil {
		return clientError
	}
	s.client = client
	return nil
}

func (s *RunStatusAction) SetupLogs() *model.Error {
	logFile := slug.Make(fmt.Sprintf("events-%s-%d", s.migration.Name, s.migration.ID))
	path := fmt.Sprintf("%s/%s.log", "logs/migration", logFile)
	log, logError := events.NewLogController(path)
	if logError != nil {
		return logError
	}
	s.Log = log
	return nil
}
