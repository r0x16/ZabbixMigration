package runjob

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/action"
	sharedDomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
)

type HostDisable struct {
	run         *RunAction
	eventId     string
	SourceProxy *model.ZabbixProxy
	srcApi      sharedDomain.ZabbixConnectorProvider
}

func NewHostDisable(run *RunAction) *HostDisable {
	return &HostDisable{
		run:     run,
		eventId: fmt.Sprintf("migration-run-%d", run.Migration.ID),
		srcApi:  nil,
	}
}

func (s *HostDisable) Run() *model.Error {
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

	s.run.Log.WriteLog("Starting source host disabling")

	s.run.Bundle.ServerEvents[s.eventId] = events.NewServerEventEcho(s.eventId)
	go s.start()

	return nil
}

func (s *HostDisable) start() {
	s.registerLog("Preparing source host disabling")

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

	disableError := s.disableHosts(hosts)
	if disableError != nil {
		s.registerLog("[ERROR]" + disableError.Message)
		s.stop(false)
		return
	}

	s.stop(true)
}

func (s *HostDisable) disableHosts(hosts []*model.ZabbixHost) *model.Error {
	for _, host := range hosts {
		s.registerLog(fmt.Sprintf("Disabling host %s", host.Host))
		disableError := s.disable(host)
		if disableError != nil {
			return disableError
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (s *HostDisable) disable(host *model.ZabbixHost) *model.Error {
	s.registerLog(fmt.Sprintf("Disabling host \"%s\"", host.Host))
	if host.Status != "0" || host.Disabled == 1 {
		s.registerLog(fmt.Sprintf("Host \"%s\" already disabled, omiting", host.Host))
		return nil
	}

	disableError := s.disableHostUsingApi(host)
	if disableError != nil {
		return disableError
	}

	host.Disabled = 1
	updateError := s.run.HostRepo.Update(host)
	if updateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: updateError.Error(),
		}
	}
	return nil
}

func (s *HostDisable) disableHostUsingApi(host *model.ZabbixHost) *model.Error {
	disabled, disableError := s.srcApi.Request(s.srcApi.Body("host.update", model.ZabbixParams{
		"hostid": host.HostID,
		"status": "1",
	}))

	if disableError != nil {
		return disableError
	}

	jsonString, _ := json.MarshalIndent(disabled.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Disabled host data result: %s", jsonString))

	return nil
}

func (s *HostDisable) apiSetup() *model.Error {
	s.srcApi = zabbix.ServerConnector(&s.run.Migration.Source)
	srcConnectError := s.srcApi.Connect(s.run.Migration.Source.Username, s.run.Migration.Source.Password)
	if srcConnectError != nil {
		return srcConnectError
	}

	return nil
}

func (s *HostDisable) stop(sucess bool) {
	s.registerLog("Finishing host disabling")
	handler := s.run.Bundle.ServerEvents[s.eventId]

	s.run.Migration.IsRunning = false

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultDisabling = false
		s.run.Migration.IsDefaultDisabled = sucess
	}

	stopError := s.run.MigrationRepo.Update(s.run.Migration)
	if stopError != nil {
		s.registerLog(stopError.Error())
		handler.Broadcast(&sharedDomain.EventMessage{
			Event: "error",
			Data:  "Migration finished with errors",
		})
	}

	s.SourceProxy.IsHostDisabling = false
	s.SourceProxy.IsHostDisabled = sucess

	if s.SourceProxy.ProxyID != "0" {
		proxyStopError := s.run.ProxyRepo.Update(s.SourceProxy)

		if proxyStopError != nil {
			s.registerLog(proxyStopError.Error())
			handler.Broadcast(&sharedDomain.EventMessage{
				Event: "error",
				Data:  "Proxy finished with errors",
			})
		}
	}

	handler.Broadcast(&sharedDomain.EventMessage{
		Event: "ready",
		Data:  "Host disabling closed",
	})
	delete(s.run.Bundle.ServerEvents, s.eventId)
}

func (s *HostDisable) registerLog(log string) {
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

func (s *HostDisable) setRunning() *model.Error {
	s.run.Migration.IsRunning = true

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultDisabling = true
	}

	updateError := s.run.MigrationRepo.Update(s.run.Migration)
	if updateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: updateError.Error(),
		}
	}

	s.SourceProxy.IsHostDisabling = true

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

func (s *HostDisable) extractSourceProxy() (*model.ZabbixProxy, *model.Error) {
	proxy, proxyError := action.ExtractFormSourceProxy(s.run.Context, s.run.ProxyRepo, s.run.Migration)
	if proxyError != nil {
		return nil, proxyError
	}

	if proxy.IsHostDisabling {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Hosts in proxy already disabling",
		}
	}

	if proxy.IsHostDisabled {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Host in proxy already disabled",
		}
	}

	return proxy, nil
}
