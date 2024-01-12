package runjob

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	sharedDomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
)

type HostImport struct {
	run     *RunAction
	eventId string

	SourceProxy *model.ZabbixProxy

	srcApi sharedDomain.ZabbixConnectorProvider
}

func NewHostImport(run *RunAction) *HostImport {
	return &HostImport{
		run:     run,
		eventId: fmt.Sprintf("migration-run-%d", run.Migration.ID),
	}
}

func (s *HostImport) Run() *model.Error {
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

	setRunningError := s.setRunning()
	if setRunningError != nil {
		return setRunningError
	}

	s.run.Log.WriteLog("Starting host import")

	s.run.Bundle.ServerEvents[s.eventId] = events.NewServerEventEcho(s.eventId)
	fmt.Println(s.run.Bundle.ServerEvents)
	go s.start()

	return nil

}

func (s *HostImport) start() {
	s.registerLog("Starting host data import")

	apiSetupError := s.apiSetup()
	if apiSetupError != nil {
		s.registerLog(apiSetupError.Message)
		s.stop(false)
		return
	}

	extractError := s.extractAndStore()
	if extractError != nil {
		s.registerLog(extractError.Message)
		s.stop(false)
		return
	}

	s.stop(true)

}

func (s *HostImport) stop(sucess bool) {
	s.registerLog("Finishing host import")
	handler := s.run.Bundle.ServerEvents[s.eventId]

	s.run.Migration.IsRunning = false
	stopError := s.run.MigrationRepo.Update(s.run.Migration)
	if stopError != nil {
		s.registerLog(stopError.Error())
		handler.Broadcast(&sharedDomain.EventMessage{
			Event: "error",
			Data:  "Migration finished with errors",
		})
	}

	s.SourceProxy.IsHostImporting = false
	s.SourceProxy.IsHostImported = sucess
	proxyStopError := s.run.ProxyRepo.Update(s.SourceProxy)

	if proxyStopError != nil {
		s.registerLog(proxyStopError.Error())
		handler.Broadcast(&sharedDomain.EventMessage{
			Event: "error",
			Data:  "Proxy finished with errors",
		})
	}

	handler.Broadcast(&sharedDomain.EventMessage{
		Event: "ready",
		Data:  "Host import closed",
	})
	delete(s.run.Bundle.ServerEvents, s.eventId)
}

func (s *HostImport) extractAndStore() *model.Error {
	s.registerLog(fmt.Sprintf("Extracting hosts from server: %s, proxy: %s", s.SourceProxy.ZabbixServer.Name, s.SourceProxy.Host))

	templates, err := s.getHostsFromApi()
	if err != nil {
		return err
	}

	s.setHostsMigration(templates)

	storeError := s.run.HostRepo.MultipleStore(templates)
	if storeError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: storeError.Error(),
		}
	}

	return nil

}

func (s *HostImport) getHostsFromApi() ([]*model.ZabbixHost, *model.Error) {
	hosts, err := s.srcApi.Request(s.srcApi.Body("host.get", model.ZabbixParams{
		"output":   []string{"hostid", "host", "proxy_hostid", "status"},
		"proxyids": []string{s.SourceProxy.ProxyID},
	}))

	if err != nil {
		return nil, err
	}

	hostList, hostListError := s.decodeHosts(hosts)
	if hostListError != nil {
		fmt.Println(hostListError)
		return nil, hostListError
	}

	fmt.Println("Decoded: ", hostList)

	return hostList, nil
}

// Decode ZabbixHost from API
func (s *HostImport) decodeHosts(hosts *model.ZabbixResponse) ([]*model.ZabbixHost, *model.Error) {
	var hostList []*model.ZabbixHost
	decodeError := json.Unmarshal([]byte(hosts.RawResult), &hostList)

	if decodeError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: decodeError.Error(),
		}
	}

	return hostList, nil
}

func (s *HostImport) setHostsMigration(hosts []*model.ZabbixHost) {
	for _, host := range hosts {
		host.Migration = s.run.Migration
	}
}

func (s *HostImport) extractSourceProxy() (*model.ZabbixProxy, *model.Error) {
	proxy, proxyError := action.ExtractFormSourceProxy(s.run.Context, s.run.ProxyRepo, s.run.Migration)
	if proxyError != nil {
		return nil, proxyError
	}

	if proxy.IsHostImporting {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Another import running",
		}
	}

	if proxy.IsHostImported {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Proxy already imported",
		}
	}

	return proxy, nil
}

func (s *HostImport) setRunning() *model.Error {
	s.run.Migration.IsRunning = true

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultHostImporting = true
	}

	updateError := s.run.MigrationRepo.Update(s.run.Migration)
	if updateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: updateError.Error(),
		}
	}

	s.SourceProxy.IsHostImporting = true
	proxyUpdateError := s.run.ProxyRepo.Update(s.SourceProxy)
	if proxyUpdateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: proxyUpdateError.Error(),
		}
	}

	return nil
}

func (s *HostImport) apiSetup() *model.Error {
	s.srcApi = zabbix.ServerConnector(&s.run.Migration.Source)
	srcConnectError := s.srcApi.Connect(s.run.Migration.Source.Username, s.run.Migration.Source.Password)
	if srcConnectError != nil {
		return srcConnectError
	}

	return nil
}

func (s *HostImport) registerLog(log string) {
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
