package tplmap

import (
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
)

type MappingBase struct {
	bundle    *drivers.ApplicationBundle
	migration *model.Migration

	SourceTemplates      []*model.ZabbixTemplate
	DestinationTemplates []*model.ZabbixTemplate

	BaseTemplateMap map[uint]uint
}

func SetupBaseMapping(bundle *drivers.ApplicationBundle, migration *model.Migration) (*MappingBase, *model.Error) {
	mapping := &MappingBase{
		bundle:          bundle,
		migration:       migration,
		BaseTemplateMap: make(map[uint]uint),
	}

	templatesError := mapping.setupTemplates()
	if templatesError != nil {
		return nil, templatesError
	}

	mapping.setupBaseTemplateMap()

	return mapping, nil
}

func (m *MappingBase) setupTemplates() *model.Error {
	sourceTemplates, err := m.getImportedTemplates(&m.migration.Source)
	if err != nil {
		return err
	}
	m.SourceTemplates = sourceTemplates

	destinationTemplates, err := m.getImportedTemplates(&m.migration.Destination)
	if err != nil {
		return err
	}
	m.DestinationTemplates = destinationTemplates

	return nil
}

func (m *MappingBase) getImportedTemplates(server *model.ZabbixServer) ([]*model.ZabbixTemplate, *model.Error) {
	repo := repository.NewZabbixTemplateRepository(m.bundle.Database.Connection)

	templates, err := repo.GetByMigrationAndServer(m.migration.ID, server.ID)
	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return templates, nil
}

func (m *MappingBase) setupBaseTemplateMap() {
	for _, sourceTemplate := range m.SourceTemplates {
		for _, destinationTemplate := range m.DestinationTemplates {
			if sourceTemplate.Host == destinationTemplate.Host || sourceTemplate.RemoteFound == destinationTemplate.Host {
				m.BaseTemplateMap[sourceTemplate.ID] = destinationTemplate.ID
			}
		}
	}
}
