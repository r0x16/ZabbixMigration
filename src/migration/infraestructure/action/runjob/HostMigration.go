package runjob

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
	sharedDomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
)

type HostMigration struct {
	run      *RunAction
	eventId  string
	migrated []string
	srcApi   sharedDomain.ZabbixConnectorProvider
	destApi  sharedDomain.ZabbixConnectorProvider
}

func NewHostMigration(run *RunAction) *HostMigration {
	return &HostMigration{
		run:     run,
		eventId: fmt.Sprintf("migration-run-%d", run.Migration.ID),
	}
}

func (s *HostMigration) Run() *model.Error {
	return nil
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
