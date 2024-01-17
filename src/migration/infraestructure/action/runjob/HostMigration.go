package runjob

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	sharedDomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
)

type HostMigration struct {
	run     *RunAction
	eventId string
	srcApi  sharedDomain.ZabbixConnectorProvider
	destApi sharedDomain.ZabbixConnectorProvider

	SourceProxy      *model.ZabbixProxy
	DestinationProxy *model.ZabbixProxy
}

func NewHostMigration(run *RunAction) *HostMigration {
	return &HostMigration{
		run:     run,
		eventId: fmt.Sprintf("migration-run-%d", run.Migration.ID),
	}
}

func (s *HostMigration) Run() *model.Error {
	_, started := s.run.Bundle.ServerEvents[s.eventId]
	if started {
		return &model.Error{
			Code:    http.StatusForbidden,
			Message: "Migration already running",
		}
	}

	var proxyError *model.Error
	s.SourceProxy, proxyError = s.extractSourceProxy()
	if proxyError != nil {
		return proxyError
	}

	var destinationProxyError *model.Error
	s.DestinationProxy, destinationProxyError = s.extractDestinationProxy()
	if destinationProxyError != nil {
		return destinationProxyError
	}

	setRunningError := s.setRunning()
	if setRunningError != nil {
		return setRunningError
	}

	s.run.Log.WriteLog("Starting host migration")

	s.run.Bundle.ServerEvents[s.eventId] = events.NewServerEventEcho(s.eventId)
	go s.start()

	return nil
}

func (s *HostMigration) start() {
	s.registerLog("Starting host configuration import")

	apiSetupError := s.apiSetup()
	if apiSetupError != nil {
		s.registerLog(apiSetupError.Message)
		s.stop(false)
		return
	}

	hosts, hostsError := s.run.HostRepo.FindByMigrationAndProxy(s.run.Migration, s.SourceProxy)
	if hostsError != nil {
		s.registerLog(hostsError.Error())
		s.stop(false)
		return
	}

	for _, host := range hosts {
		s.registerLog(fmt.Sprintf("Migrating host %s", host.Host))
		migrateError := s.migrate(host)
		if migrateError != nil {
			s.registerLog("[ERROR]" + migrateError.Message)
			s.stop(false)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	s.stop(true)
}

func (s *HostMigration) migrate(host *model.ZabbixHost) *model.Error {
	// Get updated source host with template data
	sourceHost, sourceHostError := s.getSourceHost(host.HostID)
	if sourceHostError != nil {
		s.registerLog(fmt.Sprintf("Error importing host: \"%s\"", host.Host))
		return sourceHostError
	}

	s.registerLog(fmt.Sprintf("Host \"%s\" updated data imported Succesful", sourceHost.Host))

	// Check already exists
	if s.checkAlreadyExists(sourceHost.Host) {
		s.registerLog(fmt.Sprintf("[WARNING] Host: \"%s\" already exists in %s", sourceHost.Host, s.run.Migration.Destination.Name))
		return nil
	}

	// Extract host configuration
	configuration, configError := s.getSourceHostConfiguration(sourceHost.HostID)
	if configError != nil {
		s.registerLog(fmt.Sprintf("Error extracting template configuration: \"%s\"", sourceHost.Host))
		s.registerLog(fmt.Sprintf("Host configuration: \"%s\"", configuration))
		return configError
	}

	// Override template destination data
	configuration, mappingError := s.mapDestinationTemplate(sourceHost.Templates, configuration)
	if mappingError != nil {
		s.registerLog(fmt.Sprintf("Error mapping template configuration: \"%s\"", sourceHost.Host))
		s.registerLog(fmt.Sprintf("Host configuration: \"%s\"", configuration))
		return mappingError
	}

	// Override proxy configuration
	configuration = s.mapDestinationProxy(configuration)

	// Add mighration groups
	configuration = s.addMigrationGroups(configuration)

	s.registerLog(fmt.Sprintf("Final host configuration: \"%s\"", configuration))

	// Create host
	s.registerLog(fmt.Sprintf("Creating host: \"%s\" in \"%s\"", sourceHost.Host, s.run.Migration.Destination.Name))
	createError := s.createHost(configuration)
	if createError != nil {
		s.registerLog(fmt.Sprintf("Error creating host: \"%s\"", sourceHost.Host))
		return createError
	}

	return nil
}

func (s *HostMigration) createHost(configuration string) *model.Error {
	configuration = s.debugConfiguration(configuration)

	imported, err := s.destApi.Request(s.destApi.Body("configuration.import", model.ZabbixParams{
		"format": "xml",
		"rules": model.ZabbixParams{
			"hosts":           model.ZabbixParams{"createMissing": true},
			"templateLinkage": model.ZabbixParams{"createMissing": true},
			"discoveryRules":  model.ZabbixParams{"createMissing": true},
			"graphs":          model.ZabbixParams{"createMissing": true},
			"host_groups":     model.ZabbixParams{"createMissing": true},
			"httptests":       model.ZabbixParams{"createMissing": true},
			"images":          model.ZabbixParams{"createMissing": true},
			"items":           model.ZabbixParams{"createMissing": true},
			"maps":            model.ZabbixParams{"createMissing": true},
			"mediaTypes":      model.ZabbixParams{"createMissing": true},
			"triggers":        model.ZabbixParams{"createMissing": true},
			"valueMaps":       model.ZabbixParams{"createMissing": true},
		},
		"source": configuration,
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
			Message: "Error writing host on destination",
		}
	}

	jsonString, _ := json.MarshalIndent(imported.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Migrated host data result: %s", jsonString))

	return nil
}

func (s *HostMigration) debugConfiguration(configuration string) string {
	// deleting <request_method>0</request_method> from configuration
	// because it's not supported by zabbix 6.4+
	configuration = strings.ReplaceAll(configuration, "<request_method>1</request_method>", "<request_method>0</request_method>")

	return configuration
}

func (s *HostMigration) addMigrationGroups(configuration string) string {
	oldGroups := "</groups>"
	newGroups := fmt.Sprintf("<group><name>Source: %s</name></group>", s.run.Migration.Source.Name)
	newGroups += fmt.Sprintf("<group><name>Source Proxy: %s</name></group>", s.SourceProxy.Host)
	newGroups += "<group><name>Migrated</name></group></groups>"

	return strings.Replace(configuration, oldGroups, newGroups, -1)
}

func (s *HostMigration) mapDestinationProxy(configuration string) string {
	s.registerLog("Mapping destination proxy")
	var oldProxy string
	if s.SourceProxy.ProxyID == "0" {
		oldProxy = "<proxy/>"
	} else {
		oldProxy = fmt.Sprintf("<proxy><name>%s</name></proxy>", s.SourceProxy.Host)
	}

	newProxy := fmt.Sprintf("<proxy><name>%s</name></proxy>", s.DestinationProxy.Host)

	s.registerLog(fmt.Sprintf("Replacing proxy: \"%s\" with \"%s\"", oldProxy, newProxy))

	return strings.Replace(configuration, oldProxy, newProxy, -1)
}

func (s *HostMigration) mapDestinationTemplate(templates []*model.ZabbixTemplate, configuration string) (string, *model.Error) {
	s.registerLog("Mapping destination template")
	var replaceError *model.Error
	for _, template := range templates {
		configuration, replaceError = s.replaceTemplateConfiguration(template, configuration)
		if replaceError != nil {
			return "", replaceError
		}
	}

	return configuration, nil
}

func (s *HostMigration) replaceTemplateConfiguration(template *model.ZabbixTemplate, configuration string) (string, *model.Error) {
	srcTemplate, dstTemplateError := s.run.TemplateRepo.GetByTemplateIdAndServer(template.Templateid, s.run.Migration.Source.ID, s.run.Migration.ID)

	if dstTemplateError != nil || srcTemplate == nil {
		return "", &model.Error{
			Code:    http.StatusInternalServerError,
			Message: dstTemplateError.Error(),
		}
	}

	oldTemplate := fmt.Sprintf("<template><name>%s</name></template>", srcTemplate.Host)
	newTemplate := fmt.Sprintf("<template><name>%s</name></template>", srcTemplate.SourceMapping.DestinationTemplate.Host)

	s.registerLog(fmt.Sprintf("Replacing template: \"%s\" with \"%s\"", oldTemplate, newTemplate))

	return strings.Replace(configuration, oldTemplate, newTemplate, -1), nil
}

func (s *HostMigration) getSourceHost(hostid string) (*model.ZabbixHost, *model.Error) {
	hosts, err := s.srcApi.Request(s.srcApi.Body("host.get", model.ZabbixParams{
		"hostids":               hostid,
		"output":                []string{"hostid", "host"},
		"selectParentTemplates": []string{"templateid", "host"},
	}))
	if err != nil {
		return nil, err
	}

	jsonString, _ := json.MarshalIndent(hosts.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Imported host data: %s", jsonString))

	hostList, err := s.decode(hosts)
	if err != nil {
		return nil, err
	}

	if len(hostList) == 0 {
		return nil, &model.Error{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("hostId \"%s\" not found in Server: \"%s\"", hostid, s.run.Migration.Source.Name),
		}
	}

	return hostList[0], nil
}

func (s *HostMigration) getSourceHostConfiguration(hostid string) (string, *model.Error) {
	s.registerLog("Extracting host configuration")

	exportedConfiguration, err := s.srcApi.Request(s.srcApi.Body("configuration.export", model.ZabbixParams{
		"options": model.ZabbixParams{
			"hosts": []string{hostid},
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
			Message: "Error parsing host configuration",
		}
	}

	s.registerLog("Configuration extracted successfully")

	return stringConfiguration, nil
}

func (s *HostMigration) checkAlreadyExists(hostName string) bool {
	result, err := s.destApi.Request(s.destApi.Body("host.get", model.ZabbixParams{
		"filter": model.ZabbixParams{
			"host": hostName,
		},
		"output": []string{"hostid", "host"},
	}))
	if err != nil {
		return false
	}

	jsonString, _ := json.MarshalIndent(result.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Checking if destination already exists: %s", jsonString))

	resultList, err := s.decode(result)
	if err != nil {
		return false
	}

	return len(resultList) > 0
}

func (s *HostMigration) decode(result *model.ZabbixResponse) ([]*model.ZabbixHost, *model.Error) {
	var hostList []*model.ZabbixHost
	decodeError := json.Unmarshal([]byte(result.RawResult), &hostList)

	if decodeError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: decodeError.Error(),
		}
	}

	return hostList, nil
}

func (s *HostMigration) extractSourceProxy() (*model.ZabbixProxy, *model.Error) {
	proxy, proxyError := action.ExtractFormSourceProxy(s.run.Context, s.run.ProxyRepo, s.run.Migration)
	if proxyError != nil {
		return nil, proxyError
	}

	if proxy.IsHostsRunning {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Migration already running",
		}
	}

	if proxy.IsHostSuccessful {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Proxy already migrated",
		}
	}

	return proxy, nil
}

func (s *HostMigration) extractDestinationProxy() (*model.ZabbixProxy, *model.Error) {
	proxy := s.SourceProxy

	if proxy.ProxyID == "0" {
		return s.run.Migration.DefaultProxy, nil
	}

	return proxy.SourceMapping.DestinationProxy, nil
}

func (s *HostMigration) setRunning() *model.Error {
	s.run.Migration.IsRunning = true

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultRunning = true
	}

	updateError := s.run.MigrationRepo.Update(s.run.Migration)

	if updateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: updateError.Error(),
		}
	}

	s.SourceProxy.IsHostsRunning = true
	if s.SourceProxy.ProxyID == "0" {
		return nil
	}

	proxyUpdateError := s.run.ProxyRepo.Update(s.SourceProxy)
	if proxyUpdateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: proxyUpdateError.Error(),
		}
	}

	return nil
}

func (s *HostMigration) stop(sucess bool) {
	s.registerLog("Finishing host migration")

	if !sucess {
		s.registerLog("Migration finished with error, check the logs for more details")
	}

	s.run.Migration.IsRunning = false

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultRunning = false
		s.run.Migration.IsDefaultSuccessful = sucess
	}

	stopError := s.run.MigrationRepo.Update(s.run.Migration)
	if stopError != nil {
		s.registerLog(stopError.Error())
	}

	s.SourceProxy.IsHostsRunning = false
	s.SourceProxy.IsHostSuccessful = sucess

	if s.SourceProxy.ProxyID != "0" {

		proxyStopError := s.run.ProxyRepo.Update(s.SourceProxy)

		if proxyStopError != nil {
			s.registerLog(proxyStopError.Error())
		}
	}

	s.run.Bundle.ServerEvents[s.eventId].Broadcast(&sharedDomain.EventMessage{
		Event: "ready",
		Data:  "Host migration closed",
	})
	delete(s.run.Bundle.ServerEvents, s.eventId)
}

func (s *HostMigration) apiSetup() *model.Error {
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

func (s *HostMigration) registerLog(log string) {
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

func (s *HostMigration) GetMigrationInfo() (*domain.HostMigrationInfo, *model.Error) {
	proxies, proxiesError := s.run.ProxyRepo.GetByServerWithSourceMappings(s.run.Migration.ID, s.run.Migration.SourceID)
	if proxiesError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: proxiesError.Error(),
		}
	}

	return &domain.HostMigrationInfo{
		TotalCount: len(proxies),
		Proxies:    proxies,
	}, nil

}
