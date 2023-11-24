package runjob

import (
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
)

type TemplateMigration struct {
	run *RunAction
}

func NewTemplateMigration(run *RunAction) *TemplateMigration {
	return &TemplateMigration{
		run: run,
	}
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
