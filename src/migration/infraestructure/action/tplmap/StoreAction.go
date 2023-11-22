package tplmap

import (
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"github.com/labstack/echo/v4"
)

type StoreAction struct {
	c            echo.Context
	bundle       *drivers.ApplicationBundle
	migration    *model.Migration
	templateRepo *repository.ZabbixTemplateRepository
}

func Store(migration *model.Migration, c echo.Context, bundle *drivers.ApplicationBundle) *model.Error {
	store := &StoreAction{
		c:            c,
		bundle:       bundle,
		migration:    migration,
		templateRepo: repository.NewZabbixTemplateRepository(bundle.Database.Connection),
	}

	body, mappingError := store.bindMappingBody()
	if mappingError != nil {
		return mappingError
	}

	validationError := store.validateMapping(body)
	if validationError != nil {
		return validationError
	}

	storeError := store.storeMapping(body)
	if storeError != nil {
		return storeError
	}

	markError := store.updateMigrationAsMapped()
	if markError != nil {
		return markError
	}

	return nil
}

func (s *StoreAction) bindMappingBody() (*domain.TemplateMappingBody, *model.Error) {
	var body domain.TemplateMappingBody
	bindingError := s.c.Bind(&body)
	if bindingError != nil {
		return nil, &model.Error{
			Code:    http.StatusBadRequest,
			Message: bindingError.Error(),
		}
	}

	var importError *model.Error

	body.ImportedSourceTemplates, importError = s.getImportedTemplates(&s.migration.Source)
	if importError != nil {
		return nil, importError
	}

	body.ImportedDestinationTemplates, importError = s.getImportedTemplates(&s.migration.Destination)
	if importError != nil {
		return nil, importError
	}

	return &body, nil
}

func (s *StoreAction) getImportedTemplates(server *model.ZabbixServer) ([]*model.ZabbixTemplate, *model.Error) {

	templates, err := s.templateRepo.GetByMigrationAndServer(s.migration.ID, server.ID)
	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return templates, nil
}

func (s *StoreAction) validateMapping(body *domain.TemplateMappingBody) *model.Error {
	if len(body.SourceTemplates) != len(body.DestinationTemplates) {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Source and destination templates must be the same length",
		}
	}

	srcTemplate, proxyPresentError := s.validateProxyPresent(body.SourceTemplates, body.ImportedSourceTemplates)
	if !proxyPresentError {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Source template %d not found", srcTemplate),
		}
	}

	dstTemplate, proxyPresentError := s.validateProxyPresent(s.clearEmptyMappings(body.DestinationTemplates), body.ImportedDestinationTemplates)
	if !proxyPresentError {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Destination template %d not found", dstTemplate),
		}
	}

	return nil
}

func (s *StoreAction) clearEmptyMappings(templates []uint) []uint {
	var result []uint
	for _, template := range templates {
		if template != 0 {
			result = append(result, template)
		}
	}
	return result
}

func (s *StoreAction) validateProxyPresent(templateIds []uint, templates []*model.ZabbixTemplate) (uint, bool) {
	for _, templateId := range templateIds {
		templatePresent := false

		for _, template := range templates {
			if template.ID == templateId {
				templatePresent = true
				break
			}
		}

		if !templatePresent {
			return templateId, false
		}
	}

	return 0, true
}

func (s *StoreAction) storeMapping(body *domain.TemplateMappingBody) *model.Error {
	for index, sourceTemplate := range body.SourceTemplates {

		destinationProxy := body.DestinationTemplates[index]

		if destinationProxy == 0 {
			continue
		}

		templateMap := &model.ZabbixTemplateMapping{
			SourceTemplateID:      sourceTemplate,
			DestinationTemplateID: destinationProxy,
		}

		err := s.templateRepo.StoreMapping(templateMap)

		if err != nil {
			return &model.Error{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Error storing mapping: %v, error: %s", destinationProxy, err.Error()),
			}
		}
	}
	return nil
}

func (s *StoreAction) updateMigrationAsMapped() *model.Error {
	s.migration.HasTemplateBindings = true
	repo := repository.NewMigrationRepository(s.bundle.Database.Connection)
	err := repo.Update(s.migration)
	if err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}
