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
