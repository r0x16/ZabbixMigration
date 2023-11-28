package runjob

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
)

type TemplateMigration struct {
	run     *RunAction
	eventId string
}

func NewTemplateMigration(run *RunAction) *TemplateMigration {
	return &TemplateMigration{
		run:     run,
		eventId: fmt.Sprintf("template-migration-%d", run.Migration.ID),
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
			Message: "Template migration already running",
		}
	}

	s.run.Log.WriteLog("Starting template migration")

	s.run.Bundle.ServerEvents[s.eventId] = events.NewServerEventEcho(s.eventId)
	go s.start()

	return nil
}

func (s *TemplateMigration) start() {

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
