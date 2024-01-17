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

type HostRollback struct {
	run              *RunAction
	eventId          string
	SourceProxy      *model.ZabbixProxy
	srcApi           sharedDomain.ZabbixConnectorProvider
	DestinationProxy *model.ZabbixProxy
	dstApi           sharedDomain.ZabbixConnectorProvider
}

func NewHostRollback(run *RunAction) *HostRollback {
	return &HostRollback{
		run:     run,
		eventId: fmt.Sprintf("migration-run-%d", run.Migration.ID),
		srcApi:  nil,
		dstApi:  nil,
	}
}

func (s *HostRollback) Run() *model.Error {
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

	s.run.Log.WriteLog("Starting host migration rollback")

	s.run.Bundle.ServerEvents[s.eventId] = events.NewServerEventEcho(s.eventId)
	go s.start()

	return nil
}

func (s *HostRollback) start() {
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

	mapError := s.mapDestinationHost(hosts)
	if mapError != nil {
		s.registerLog("[ERROR]" + mapError.Message)
		s.stop(false)
		return
	}

	rollbackError := s.rollbackHosts(hosts)
	if rollbackError != nil {
		s.registerLog("[ERROR]" + rollbackError.Message)
		s.stop(false)
		return
	}

	rollbackActionsError := s.rollbackActions()
	if rollbackActionsError != nil {
		s.registerLog("[ERROR]" + rollbackActionsError.Message)
		s.stop(false)
		return
	}

	s.registerLog("Host rollback finished")
	s.stop(true)

}

func (s *HostRollback) rollbackHosts(hosts []*model.ZabbixHost) *model.Error {
	for _, host := range hosts {
		rollbackError := s.rollback(host)
		if rollbackError != nil {
			return rollbackError
		}
	}
	return nil
}

func (s *HostRollback) rollbackActions() *model.Error {
	s.registerLog("Rollback migration actions")

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultDisabled = false
		s.run.Migration.IsDefaultSuccessful = false

		stopError := s.run.MigrationRepo.Update(s.run.Migration)
		if stopError != nil {
			s.registerLog(stopError.Error())
		}
	}

	s.SourceProxy.IsHostDisabled = false
	s.SourceProxy.IsHostSuccessful = false

	if s.SourceProxy.ProxyID != "0" {

		proxyStopError := s.run.ProxyRepo.Update(s.SourceProxy)

		if proxyStopError != nil {
			s.registerLog(proxyStopError.Error())
		}
	}
	return nil
}

func (s *HostRollback) rollback(host *model.ZabbixHost) *model.Error {
	s.registerLog(fmt.Sprintf("Rollback host: %s", host.Host))

	deleteError := s.deleteDestinationHost(host)
	if deleteError != nil {
		return deleteError
	}

	if host.Disabled == 1 {
		s.registerLog(fmt.Sprintf("Host %s is disabled in source, enabling...", host.Host))
		enableError := s.enableSourceHost(host)
		if enableError != nil {
			return enableError
		}
	}
	return nil
}

func (s *HostRollback) deleteDestinationHost(host *model.ZabbixHost) *model.Error {
	s.registerLog(fmt.Sprintf("Deleting host \"%s\" from \"%s\"", host.Host, s.run.Migration.Destination.Name))
	disabled, disableError := s.dstApi.ArrayRequest(s.dstApi.ArrayBody("host.delete", []string{
		host.DstHostID,
	}))

	if disableError != nil {
		return disableError
	}

	jsonString, _ := json.MarshalIndent(disabled.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Deleted host data result: %s", jsonString))

	return nil
}

func (s *HostRollback) enableSourceHost(host *model.ZabbixHost) *model.Error {
	s.registerLog(fmt.Sprintf("Enabling host \"%s\" from \"%s\"", host.Host, s.run.Migration.Source.Name))
	enabled, enableError := s.srcApi.Request(s.srcApi.Body("host.update", model.ZabbixParams{
		"hostid": host.HostID,
		"status": "0",
	}))

	if enableError != nil {
		return enableError
	}

	jsonString, _ := json.MarshalIndent(enabled.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Enabled host data result: %s", jsonString))

	return nil
}

func (s *HostRollback) mapDestinationHost(hosts []*model.ZabbixHost) *model.Error {
	for _, host := range hosts {
		mapError := s.mapDestination(host)
		if mapError != nil {
			return mapError
		}
	}
	return nil
}

func (s *HostRollback) mapDestination(host *model.ZabbixHost) *model.Error {
	hosts, hostError := s.dstApi.Request(s.dstApi.Body("host.get", model.ZabbixParams{
		"filter": model.ZabbixParams{
			"host": host.Host,
		},
		"output": []string{"hostid", "host"},
	}))
	if hostError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: hostError.Error(),
		}
	}

	jsonString, _ := json.MarshalIndent(hosts.Result, "", "\t")
	s.registerLog(fmt.Sprintf("Checking if destination exists: %s", jsonString))

	hostList, decodeError := s.decode(hosts)
	if decodeError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: decodeError.Error(),
		}
	}

	if len(hostList) == 0 {
		return &model.Error{
			Code:    http.StatusNotFound,
			Message: "Host not found in destination",
		}
	}

	host.DstHostID = hostList[0].HostID

	return nil
}

func (s *HostRollback) decode(hosts *model.ZabbixResponse) ([]*model.ZabbixHost, *model.Error) {
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

func (s *HostRollback) apiSetup() *model.Error {
	s.srcApi = zabbix.ServerConnector(&s.run.Migration.Source)
	srcConnectError := s.srcApi.Connect(s.run.Migration.Source.Username, s.run.Migration.Source.Password)
	if srcConnectError != nil {
		return srcConnectError
	}

	s.dstApi = zabbix.ServerConnector(&s.run.Migration.Destination)
	destConnectError := s.dstApi.Connect(s.run.Migration.Destination.Username, s.run.Migration.Destination.Password)
	if destConnectError != nil {
		return destConnectError
	}

	return nil
}

func (s *HostRollback) registerLog(log string) {
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

func (s *HostRollback) stop(sucess bool) {
	s.registerLog("Finishing host migration rollback")

	if !sucess {
		s.registerLog("Migration finished with error, check the logs for more details")
	}

	s.run.Migration.IsRunning = false

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultRollingBack = false
	}

	stopError := s.run.MigrationRepo.Update(s.run.Migration)
	if stopError != nil {
		s.registerLog(stopError.Error())
	}

	s.SourceProxy.IsRollingBack = false

	if s.SourceProxy.ProxyID != "0" {

		proxyStopError := s.run.ProxyRepo.Update(s.SourceProxy)

		if proxyStopError != nil {
			s.registerLog(proxyStopError.Error())
		}
	}

	s.run.Bundle.ServerEvents[s.eventId].Broadcast(&sharedDomain.EventMessage{
		Event: "ready",
		Data:  "Host rollback closed",
	})
	delete(s.run.Bundle.ServerEvents, s.eventId)
}

// extractSourceProxy
func (s *HostRollback) extractSourceProxy() (*model.ZabbixProxy, *model.Error) {
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

	if !proxy.IsHostSuccessful {
		return nil, &model.Error{
			Code:    http.StatusForbidden,
			Message: "Nothing to rollback",
		}
	}

	return proxy, nil
}

// extractDestinationProxy
func (s *HostRollback) extractDestinationProxy() (*model.ZabbixProxy, *model.Error) {
	proxy := s.SourceProxy

	if proxy.ProxyID == "0" {
		return s.run.Migration.DefaultProxy, nil
	}

	return proxy.SourceMapping.DestinationProxy, nil
}

// setRunning
func (s *HostRollback) setRunning() *model.Error {
	s.run.Migration.IsRunning = true

	if s.SourceProxy.ProxyID == "0" {
		s.run.Migration.IsDefaultRollingBack = true
	}

	updateError := s.run.MigrationRepo.Update(s.run.Migration)

	if updateError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: updateError.Error(),
		}
	}

	s.SourceProxy.IsRollingBack = true
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
