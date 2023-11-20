package tplmap

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ImportStatusAction struct {
	c         echo.Context
	bundle    *drivers.ApplicationBundle
	migration *model.Migration
	eventId   string
	clientId  string
	client    events.EventClientEcho
}

func ImportStatus(c echo.Context, bundle *drivers.ApplicationBundle) error {
	importStatus := &ImportStatusAction{
		c:        c,
		bundle:   bundle,
		clientId: uuid.New().String(),
	}

	clientError := importStatus.SetupClient()
	if clientError != nil {
		return echo.NewHTTPError(clientError.Code, clientError.Message)
	}

	migrationError := importStatus.SetupMigration()
	if migrationError != nil {
		return echo.NewHTTPError(migrationError.Code, migrationError.Message)
	}

	if importStatus.migration.IsTemplateImported {
		importStatus.client.Close()
		return c.NoContent(http.StatusOK)
	}

	importStatus.RunClient()
	importStatus.StopClient()
	return c.NoContent(http.StatusOK)
}

func (s *ImportStatusAction) SetupMigration() *model.Error {
	migration, importError := CheckTemplateImport(s.c, s.bundle)
	if importError != nil {
		return importError
	}
	s.migration = migration
	s.eventId = fmt.Sprintf("template-import-%d", migration.ID)
	return nil
}

func (s *ImportStatusAction) SetupClient() *model.Error {
	client := events.NewEventClientEcho(s.clientId, s.c)
	clientError := client.Setup()
	if clientError != nil {
		return clientError
	}
	s.client = client
	return nil
}

func (s *ImportStatusAction) RunClient() {
	s.bundle.ServerEvents[s.eventId].Subscribe(s.client)
	s.client.Online()
}

func (s *ImportStatusAction) StopClient() {
	eventHandler, active := s.bundle.ServerEvents[s.eventId]
	if active {
		eventHandler.Unsubscribe(s.client)
	}
}
