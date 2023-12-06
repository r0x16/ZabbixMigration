package tplmap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	sharedDomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
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

	srcApi sharedDomain.ZabbixConnectorProvider
	dstApi sharedDomain.ZabbixConnectorProvider
}

type TemplateConfigMigrationCheck map[string]map[string][]map[string]map[string]interface{}

var templateMapMutex sync.Mutex

func CheckTemplateImport(c echo.Context, bundle *drivers.ApplicationBundle) (*model.Migration, *model.Error) {
	templateMapMutex.Lock()
	defer templateMapMutex.Unlock()

	migration, err := action.GetMigrationFromParam(c, bundle)
	if err != nil {
		return nil, err
	}

	if !migration.IsTemplateImported {
		startTemplateImport(c, bundle, migration)
	}

	return migration, nil
}

func startTemplateImport(c echo.Context, bundle *drivers.ApplicationBundle, migration *model.Migration) *model.Error {
	templateImport := &TemplateImport{
		c:         c,
		bundle:    bundle,
		migration: migration,
		event_id:  fmt.Sprintf("template-import-%d", migration.ID),
	}

	apiSetupError := templateImport.apiSetup()
	if apiSetupError != nil {
		return apiSetupError
	}

	templateImport.start()

	return nil
}

func (ti *TemplateImport) start() {
	_, started := ti.bundle.ServerEvents[ti.event_id]
	if !started {
		ti.bundle.ServerEvents[ti.event_id] = events.NewServerEventEcho(ti.event_id)
		go ti.importTemplates()
	}
}

func (ti *TemplateImport) apiSetup() *model.Error {
	ti.srcApi = zabbix.ServerConnector(&ti.migration.Source)
	srcConnectError := ti.srcApi.Connect(ti.migration.Source.Username, ti.migration.Source.Password)
	if srcConnectError != nil {
		return srcConnectError
	}

	ti.dstApi = zabbix.ServerConnector(&ti.migration.Destination)
	dstConnectError := ti.dstApi.Connect(ti.migration.Destination.Username, ti.migration.Destination.Password)
	if dstConnectError != nil {
		return dstConnectError
	}

	return nil
}

func (ti *TemplateImport) importTemplates() {
	eventHandler := ti.bundle.ServerEvents[ti.event_id]

	ti.extractAndStore(&ti.migration.Source, true, ti.srcApi)
	ti.extractAndStore(&ti.migration.Destination, false, ti.dstApi)

	closeError := ti.closeImport(eventHandler)

	if closeError != nil {
		eventHandler.Broadcast(&sharedDomain.EventMessage{
			Event: "error",
			Data:  closeError,
		})
	}

	delete(ti.bundle.ServerEvents, ti.event_id)

}

func (ti *TemplateImport) extractAndStore(server *model.ZabbixServer, isSource bool, api sharedDomain.ZabbixConnectorProvider) *model.Error {
	fmt.Println("Extracting templates from server: ", server.Name)
	repo := repository.NewZabbixTemplateRepository(ti.bundle.Database.Connection)

	templates, err := ti.getTemplatesFromApi(server, api)
	if err != nil {
		return err
	}

	if isSource {
		templates = ti.clearUnusedTemplates(templates)
		remoteMapError := ti.findRemoteMap(templates)
		if remoteMapError != nil {
			return remoteMapError
		}
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

func (ti *TemplateImport) getTemplatesFromApi(server *model.ZabbixServer, api sharedDomain.ZabbixConnectorProvider) ([]*model.ZabbixTemplate, *model.Error) {

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

func (ti *TemplateImport) clearUnusedTemplates(templates []*model.ZabbixTemplate) []*model.ZabbixTemplate {
	cleanList := make([]*model.ZabbixTemplate, 0)
	for _, template := range templates {
		if template.HostCount > 0 {
			cleanList = append(cleanList, template)
		}
	}

	return cleanList
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

func (ti *TemplateImport) closeImport(eventHandler sharedDomain.ServerEventProvider) *model.Error {
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

	eventHandler.Broadcast(&sharedDomain.EventMessage{
		Event: "ready",
		Data:  "Template import finished",
	})

	return nil
}

func (ti *TemplateImport) findRemoteMap(sourceTemplates []*model.ZabbixTemplate) *model.Error {
	for _, sourceTemplate := range sourceTemplates {
		configuration, err := ti.getMigrationConfiguration(sourceTemplate.Templateid)
		if err != nil {
			return err
		}
		remoteFound, migrationError := ti.testConfigMigration(configuration)
		if migrationError != nil {
			return migrationError
		}
		sourceTemplate.RemoteFound = remoteFound
	}
	return nil
}

func (ti *TemplateImport) getMigrationConfiguration(templateId string) (string, *model.Error) {

	exportedConfiguration, err := ti.srcApi.Request(ti.srcApi.Body("configuration.export", model.ZabbixParams{
		"options": model.ZabbixParams{
			"templates": []string{templateId},
		},
		"format": "xml",
	}))
	if err != nil {
		return "", err
	}

	stringConfiguration, ok := exportedConfiguration.Result.(string)
	if !ok {
		return "", &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error parsing template configuration",
		}
	}

	return ti.debugConfiguration(stringConfiguration), nil
}

func (ti *TemplateImport) debugConfiguration(configuration string) string {
	// deleting <request_method>0</request_method> from configuration
	// because it's not supported by zabbix 6.4+
	configuration = strings.ReplaceAll(configuration, "<request_method>1</request_method>", "<request_method>0</request_method>")

	return configuration
}

func (ti *TemplateImport) testConfigMigration(configuration string) (string, *model.Error) {
	testImport, err := ti.dstApi.Request(ti.dstApi.Body("configuration.importcompare", model.ZabbixParams{
		"rules": model.ZabbixParams{
			"templates": model.ZabbixParams{"updateExisting": true},
		},
		"format": "xml",
		"source": configuration,
	}))

	if err != nil {
		return "", err
	}

	if string(testImport.RawResult) == "[]" {
		return "", nil
	}

	found := TemplateConfigMigrationCheck{}
	marshalError := json.Unmarshal(testImport.RawResult, &found)
	if marshalError != nil {
		fmt.Println(marshalError)
		return "", &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error parsing template configuration",
		}
	}

	return found["templates"]["updated"][0]["before"]["template"].(string), nil
}
