package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/repository"
	"gorm.io/gorm"
)

type ZabbixTemplateRepository struct {
	db *gorm.DB
}

var _ repository.ZabbixTemplateRepository = &ZabbixTemplateRepository{}

func NewZabbixTemplateRepository(db *gorm.DB) *ZabbixTemplateRepository {
	repo := &ZabbixTemplateRepository{db: db}
	repo.db.AutoMigrate(&model.ZabbixTemplate{})
	repo.db.AutoMigrate(&model.ZabbixTemplateMapping{})
	repo.db.AutoMigrate(&model.ZabbixParentTemplate{})
	return repo
}

// Store implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) Store(zabbixTemplate *model.ZabbixTemplate) error {
	result := r.db.Create(&zabbixTemplate)
	return result.Error
}

// GetAll implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) GetAll() ([]*model.ZabbixTemplate, error) {
	var zabbixTemplates []*model.ZabbixTemplate
	result := r.db.Find(&zabbixTemplates)
	return zabbixTemplates, result.Error
}

// GetByMigrationAndServer implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) GetByMigrationAndServer(migrationId uint, serverId uint) ([]*model.ZabbixTemplate, error) {
	var zabbixTemplates []*model.ZabbixTemplate
	result := r.db.Where("migration_id = ? AND zabbix_server_id = ?", migrationId, serverId).Find(&zabbixTemplates)
	return zabbixTemplates, result.Error
}

// GetByTemplateIdAndServer implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) GetByTemplateIdAndServer(templateId string, serverId uint) (*model.ZabbixTemplate, error) {
	var zabbixTemplate model.ZabbixTemplate
	result := r.db.Preload("DestinationMapping.DestinationTemplate").Limit(1).Find(&zabbixTemplate, "templateid = ? AND zabbix_server_id = ?", templateId, serverId)
	return &zabbixTemplate, result.Error
}

// GetWithMappingAndParents implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) GetWithMappingAndParents(migrationId uint, serverId uint) ([]*model.ZabbixTemplate, error) {
	var templates []*model.ZabbixTemplate
	result := r.db.Preload("SourceMapping").Preload("Parents").Find(&templates, "migration_id = ? AND zabbix_server_id = ?", migrationId, serverId)
	return templates, result.Error
}

// MultipleStore implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) MultipleStore(zabbixTemplates []*model.ZabbixTemplate) error {
	if len(zabbixTemplates) == 0 {
		return nil
	}
	result := r.db.Create(&zabbixTemplates)
	return result.Error
}

// StoreMapping implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) StoreMapping(mapping *model.ZabbixTemplateMapping) error {
	result := r.db.Create(&mapping)
	return result.Error
}

// GetByMigrationAndServerPreMapping implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) GetWithSourcePreMapping(migrationId uint, serverId uint) ([]*model.ZabbixTemplateMapping, error) {
	var mappings []*model.ZabbixTemplateMapping
	result := r.db.InnerJoins("SourceTemplate", r.db.Where(&model.ZabbixTemplate{
		MigrationID:    migrationId,
		ZabbixServerID: serverId,
	})).Find(&mappings)
	return mappings, result.Error
}

// GetByMigrationAndServerMapping implements repository.ZabbixTemplateRepository.
func (r *ZabbixTemplateRepository) GetWithSourceMapping(migrationId uint, serverId uint) ([]*model.ZabbixTemplateMapping, error) {
	var mappings []*model.ZabbixTemplateMapping
	result := r.db.InnerJoins("SourceTemplate", r.db.Where(&model.ZabbixTemplate{
		MigrationID:    migrationId,
		ZabbixServerID: serverId,
	})).Find(&mappings)
	return mappings, result.Error
}
