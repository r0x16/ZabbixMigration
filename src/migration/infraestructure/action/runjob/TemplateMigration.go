package runjob

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
	sharedDomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
)

type TemplateMigration struct {
	run      *RunAction
	eventId  string
	migrated []string
	srcApi   sharedDomain.ZabbixConnectorProvider
	destApi  sharedDomain.ZabbixConnectorProvider
}

func NewTemplateMigration(run *RunAction) *TemplateMigration {
	return &TemplateMigration{
		run:     run,
		eventId: fmt.Sprintf("migration-run-%d", run.Migration.ID),
	}
}

func (s *TemplateMigration) Run() *model.Error {
	setRunningError := s.setRunning()
	if setRunningError != nil {
		return setRunningError
	}

	_, started := s.run.Bundle.ServerEvents[s.eventId]
	if started {
		return &model.Error{
			Code:    http.StatusForbidden,
			Message: "Migration  already running",
		}
	}

	s.run.Log.WriteLog("Starting template migration")

	s.run.Bundle.ServerEvents[s.eventId] = events.NewServerEventEcho(s.eventId)
	go s.start()

	return nil
}

func (s *TemplateMigration) start() {
	s.registerLog("Starting template import")

	apiSetupError := s.apiSetup()
	if apiSetupError != nil {
		s.registerLog(apiSetupError.Message)
		s.stop(false)
		return
	}

	templates, templatesError := s.run.TemplateRepo.GetWithMappingAndParents(s.run.Migration.ID, s.run.Migration.SourceID)
	if templatesError != nil {
		s.registerLog(templatesError.Error())
		s.stop(false)
		return
	}

	for _, template := range templates {
		if template.SourceMapping == nil {
			s.registerLog(fmt.Sprintf("Migrating template %s", template.Name))
			migrateError := s.migrate(template.Templateid)
			if migrateError != nil {
				s.registerLog(migrateError.Message)
				s.stop(false)
				return
			}
		}
	}

	s.stop(true)
}

func (s *TemplateMigration) apiSetup() *model.Error {
	s.srcApi = zabbix.ServerConnector(&s.run.Migration.Source)
	srcConnectError := s.srcApi.Connect(s.run.Migration.Source.Username, s.run.Migration.Source.Password)
	if srcConnectError != nil {
		return srcConnectError
	}

	s.destApi = zabbix.ServerConnector(&s.run.Migration.Destination)
	destConnectError := s.destApi.Connect(s.run.Migration.Destination.Username, s.run.Migration.Destination.Password)
	if destConnectError != nil {
		return destConnectError
	}

	return nil
}

func (s *TemplateMigration) migrate(templateId string) *model.Error {
	// Import template from source
	sourceTemplate, sourceError := s.getSourceTemplate(templateId)
	if sourceError != nil {
		s.registerLog(fmt.Sprintf("Error importing templateId: \"%s\"", templateId))
		return sourceError
	}

	s.registerLog(fmt.Sprintf("Template \"%s\" imported Succesful", sourceTemplate.Host))

	// Check already migrated
	if s.checkAlreadyMigrated(templateId) {
		s.registerLog(fmt.Sprintf("Template: \"%s\" ommited", sourceTemplate.Host))
		return nil
	}

	// Check already exists
	if s.checkAlreadyExists(sourceTemplate.Host) {
		s.registerLog(fmt.Sprintf("Template: \"%s\" already exists in %s", sourceTemplate.Host, s.run.Migration.Destination.Name))
		return nil
	}

	// Extract template configuration
	configuration, configError := s.getSourceTemplateConfiguration(templateId)
	if configError != nil {
		s.registerLog(fmt.Sprintf("Error extracting template configuration: \"%s\"", sourceTemplate.Host))
		return configError
	}

	// Create parent templates
	if len(sourceTemplate.Parents) > 0 {
		s.registerLog(fmt.Sprintf("Creating parent templates for: \"%s\"", sourceTemplate.Host))
		var parentError *model.Error
		for _, parent := range sourceTemplate.Parents {
			configuration, parentError = s.createParentTemplate(strconv.Itoa(int(parent.TemplateID)), configuration)
			if parentError != nil {
				s.registerLog(fmt.Sprintf("Error creating parent template: \"%s\"", parent.Host))
				return parentError
			}
		}
	}

	// Create template
	s.registerLog(fmt.Sprintf("Creating template: \"%s\" in \"%s\"", sourceTemplate.Host, s.run.Migration.Destination.Name))
	createError := s.createTemplate(configuration)
	if createError != nil {
		s.registerLog(fmt.Sprintf("Error creating template: \"%s\"", sourceTemplate.Host))
		return createError
	}

	s.migrated = append(s.migrated, templateId)

	return nil
}

func (s *TemplateMigration) getSourceTemplate(templateId string) (*model.ZabbixTemplate, *model.Error) {

	templates, err := s.srcApi.Request(s.srcApi.Body("template.get", model.ZabbixParams{
		"templateids":           templateId,
		"output":                []string{"templateid", "host"},
		"selectParentTemplates": []string{"templateid", "host"},
	}))
	if err != nil {
		return nil, err
	}

	jsonString, _ := json.MarshalIndent(templates.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Imported template data: %s", jsonString))

	templateList, err := s.decode(templates)
	if err != nil {
		return nil, err
	}

	if len(templateList) == 0 {
		return nil, &model.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("TemplateId \"%s\" not found in Server: \"%s\"", templateId, s.run.Migration.Source.Name),
		}
	}

	return templateList[0], nil
}

func (s *TemplateMigration) decode(templates *model.ZabbixResponse) ([]*model.ZabbixTemplate, *model.Error) {
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

func (s *TemplateMigration) getSourceTemplateConfiguration(templateId string) (string, *model.Error) {
	s.registerLog("Extracting template configuration")

	exportedConfiguration, err := s.srcApi.Request(s.srcApi.Body("configuration.export", model.ZabbixParams{
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

	s.registerLog("Configuration extracted successfully")

	return stringConfiguration, nil
}

func (s *TemplateMigration) createTemplate(configuration string) *model.Error {

	imported, err := s.destApi.Request(s.destApi.Body("configuration.import", model.ZabbixParams{
		"format": "xml",
		"rules": model.ZabbixParams{
			"templates":          model.ZabbixParams{"createMissing": true},
			"discoveryRules":     model.ZabbixParams{"createMissing": true},
			"graphs":             model.ZabbixParams{"createMissing": true},
			"template_groups":    model.ZabbixParams{"createMissing": true},
			"httptests":          model.ZabbixParams{"createMissing": true},
			"images":             model.ZabbixParams{"createMissing": true},
			"items":              model.ZabbixParams{"createMissing": true},
			"maps":               model.ZabbixParams{"createMissing": true},
			"mediaTypes":         model.ZabbixParams{"createMissing": true},
			"templateDashboards": model.ZabbixParams{"createMissing": true},
			"triggers":           model.ZabbixParams{"createMissing": true},
			"valueMaps":          model.ZabbixParams{"createMissing": true},
		},
		"source": s.debugConfiguration(configuration),
	}))
	if err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("Message: %s\nData: %s", err.Data.(string), configuration),
		}
	}

	if !imported.Result.(bool) {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error writing template on destination",
		}
	}

	jsonString, _ := json.MarshalIndent(imported.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Imported template data: %s", jsonString))

	return nil
}

func (s *TemplateMigration) debugConfiguration(configuration string) string {
	// deleting <request_method>0</request_method> from configuration
	// because it's not supported by zabbix 6.4+
	configuration = strings.ReplaceAll(configuration, "<request_method>1</request_method>", "<request_method>0</request_method>")

	return configuration
}

func (s *TemplateMigration) createParentTemplate(templateId string, childConfiguration string) (string, *model.Error) {
	imported, err := s.run.TemplateRepo.GetByTemplateIdAndServer(templateId, s.run.Migration.SourceID)
	if err != nil {
		return childConfiguration, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	if imported.ID != 0 && imported.DestinationMapping != nil {
		oldName := fmt.Sprintf("<name>%s</name>", imported.Host)
		newName := fmt.Sprintf("<name>%s</name>", imported.DestinationMapping.DestinationTemplate.Host)
		return strings.Replace(childConfiguration, oldName, newName, 1), nil
	}

	s.migrate(templateId)

	return childConfiguration, nil
}

func (s *TemplateMigration) checkAlreadyMigrated(templateId string) bool {
	for _, migrated := range s.migrated {
		if migrated == templateId {
			return true
		}
	}

	return false
}

func (s *TemplateMigration) checkAlreadyExists(templateName string) bool {
	templates, err := s.destApi.Request(s.destApi.Body("template.get", model.ZabbixParams{
		"filter": model.ZabbixParams{
			"host": templateName,
		},
		"output":                []string{"templateid", "host"},
		"selectParentTemplates": []string{"templateid", "host"},
	}))
	if err != nil {
		return false
	}

	jsonString, _ := json.MarshalIndent(templates.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Checking if destination already exists: %s", jsonString))

	templateList, err := s.decode(templates)
	if err != nil {
		return false
	}

	return len(templateList) > 0
}

func (s *TemplateMigration) stop(sucess bool) {
	s.registerLog("Finishing template migration")
	s.run.Migration.IsTemplateRunning = false
	s.run.Migration.IsRunning = false
	s.run.Migration.IsTemplateSuccessful = sucess
	stopError := s.run.MigrationRepo.Update(s.run.Migration)
	if stopError != nil {
		s.registerLog(stopError.Error())
	}

	s.run.Bundle.ServerEvents[s.eventId].Broadcast(&sharedDomain.EventMessage{
		Event: "ready",
		Data:  "Template migration closed",
	})
	delete(s.run.Bundle.ServerEvents, s.eventId)
}

func (s *TemplateMigration) registerLog(log string) {
	handler := s.run.Bundle.ServerEvents[s.eventId]

	formatted, logError := s.run.Log.WriteLog(log)
	if logError != nil {
		handler.Broadcast(&sharedDomain.EventMessage{
			Event: "error",
			Data:  logError,
		})
		return
	}

	handler.Broadcast(&sharedDomain.EventMessage{
		Event: "log",
		Data:  formatted,
	})
}

func (s *TemplateMigration) setRunning() *model.Error {

	s.run.Migration.IsTemplateRunning = true
	s.run.Migration.IsRunning = true
	updateError := s.run.MigrationRepo.Update(s.run.Migration)

	if updateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: updateError.Error(),
		}
	}

	return nil
}

func (s *TemplateMigration) GetMigrationInfo() (*domain.TemplateMigrationInfo, *model.Error) {
	all, allError := s.run.TemplateRepo.GetWithMappingAndParents(s.run.Migration.ID, s.run.Migration.SourceID)
	if allError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: allError.Error(),
		}
	}

	mapped, mappedError := s.run.TemplateRepo.GetWithSourcePreMapping(s.run.Migration.ID, s.run.Migration.SourceID)
	if mappedError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: mappedError.Error(),
		}
	}

	return &domain.TemplateMigrationInfo{
		TotalCount:  len(all),
		CreateCount: len(all) - len(mapped),
		MapCount:    len(mapped),
	}, nil

}
