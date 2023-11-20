package tplmap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
	"github.com/labstack/echo/v4"
)

type TemplateImport struct {
	c         echo.Context
	bundle    *drivers.ApplicationBundle
	migration *model.Migration
	event_id  string
}

var templateMapMutex sync.Mutex

func CheckTemplateImport(c echo.Context, bundle *drivers.ApplicationBundle) (*model.Migration, *model.Error) {
	templateMapMutex.Lock()
	defer templateMapMutex.Unlock()

	migration, err := action.GetMigrationFromParam(c, bundle)
	if err != nil {
		return nil, err
	}

	if !migration.IsProxyImported {
		startTemplateImport(c, bundle, migration)
	}

	return migration, nil
}

func startTemplateImport(c echo.Context, bundle *drivers.ApplicationBundle, migration *model.Migration) {
	templateImport := &TemplateImport{
		c:         c,
		bundle:    bundle,
		migration: migration,
		event_id:  fmt.Sprintf("template-import-%d", migration.ID),
	}
	templateImport.start()
}

func (ti *TemplateImport) start() {
	_, started := ti.bundle.ServerEvents[ti.event_id]
	if !started {
		ti.bundle.ServerEvents[ti.event_id] = events.NewServerEventEcho(ti.event_id)
		go ti.importTemplates()
	}
}

func (ti *TemplateImport) importTemplates() {
	eventHandler := ti.bundle.ServerEvents[ti.event_id]

	ti.extractAndStore(&ti.migration.Source)
	ti.extractAndStore(&ti.migration.Destination)

	closeError := ti.closeImport(eventHandler)

	if closeError != nil {
		eventHandler.Broadcast(&domain.EventMessage{
			Event: "error",
			Data:  closeError,
		})
	}

	delete(ti.bundle.ServerEvents, ti.event_id)

}

func (ti *TemplateImport) extractAndStore(server *model.ZabbixServer) *model.Error {
	repo := repository.NewZabbixTemplateRepository(ti.bundle.Database.Connection)

	templates, err := ti.getTemplatesFromApi(server)
	if err != nil {
		return err
	}

	ti.setTemplatesMigration(templates)
	ti.setTemplatesServer(templates, server)

	storeError := repo.MultipleStore(templates)
	if storeError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: storeError.Error(),
		}
	}

	return nil

}

func (ti *TemplateImport) getTemplatesFromApi(server *model.ZabbixServer) ([]*model.ZabbixTemplate, *model.Error) {
	api := zabbix.ServerConnector(server)
	api.Connect(server.Username, server.Password)

	templates, err := api.Request(api.Body("template.get", model.ZabbixParams{
		"output":                "extend",
		"selectHosts":           "count",
		"selectParentTemplates": []string{"templateid", "host"},
		"selectItems":           "count",
		"selectTriggers":        "count",
		"selectGraphs":          "count",
		"selectScreens":         "count",
		"selectDiscoveries":     "count",
		"selectHttpTests":       "count",
		"selectMacros":          "count",
	}))

	if err != nil {
		return nil, err
	}

	templateList, err := ti.decode(templates)

	if err != nil {
		return nil, err
	}

	return templateList, nil
}

func (ti *TemplateImport) decode(templates *model.ZabbixResponse) ([]*model.ZabbixTemplate, *model.Error) {
	var templateList []*model.ZabbixTemplate
	decodeError := json.Unmarshal([]byte(templates.RawResult), &templateList)

	if decodeError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: decodeError.Error(),
		}
	}

	return templateList, nil
}

func (ti *TemplateImport) setTemplatesMigration(templates []*model.ZabbixTemplate) {
	for _, template := range templates {
		template.MigrationID = ti.migration.ID
	}
}

func (ti *TemplateImport) setTemplatesServer(templates []*model.ZabbixTemplate, server *model.ZabbixServer) {
	for _, template := range templates {
		template.ZabbixServerID = server.ID
	}
}

func (ti *TemplateImport) closeImport(eventHandler domain.ServerEventProvider) *model.Error {
	templateMapMutex.Lock()
	defer templateMapMutex.Unlock()

	ti.migration.IsTemplateImported = true
	repo := repository.NewMigrationRepository(ti.bundle.Database.Connection)
	if err := repo.Update(ti.migration); err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	eventHandler.Broadcast(&domain.EventMessage{
		Event: "ready",
		Data:  "Template import finished",
	})

	return nil
}
