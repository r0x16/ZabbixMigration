package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/repository"
	"gorm.io/gorm"
)

type ZabbixProxyRepository struct {
	db *gorm.DB
}

var _ repository.ZabbixProxyRepository = &ZabbixProxyRepository{}

func NewZabbixProxyRepository(db *gorm.DB) *ZabbixProxyRepository {
	repo := &ZabbixProxyRepository{db: db}
	repo.db.AutoMigrate(&model.ZabbixProxy{})
	repo.db.AutoMigrate(&model.ZabbixProxyInterface{})
	repo.db.AutoMigrate(&model.ZabbixProxyMapping{})
	return repo
}

// GetAll implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) GetAll() ([]*model.ZabbixProxy, error) {
	var zabbixProxies []*model.ZabbixProxy
	result := r.db.Joins("Interface").Find(&zabbixProxies)
	return zabbixProxies, result.Error
}

// GetById implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) GetByIdWithSourceMappings(id uint) (*model.ZabbixProxy, error) {
	var zabbixProxy model.ZabbixProxy
	result := r.db.Joins("Interface").Preload("SourceMapping.DestinationProxy").Preload("ZabbixServer").First(&zabbixProxy, id)
	return &zabbixProxy, result.Error
}

// GetByMigration implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) GetByMigrationAndServer(migrationId uint, serverId uint) ([]*model.ZabbixProxy, error) {
	var zabbixProxies []*model.ZabbixProxy
	result := r.db.Joins("Interface").Where("migration_id = ? AND zabbix_server_id = ?", migrationId, serverId).Find(&zabbixProxies)
	return zabbixProxies, result.Error
}

func (r *ZabbixProxyRepository) GetByServerWithSourceMappings(migration, serverId uint) ([]*model.ZabbixProxy, error) {
	var zabbixProxies []*model.ZabbixProxy
	result := r.db.Order("host_count asc").Preload("SourceMapping.DestinationProxy").Find(&zabbixProxies, "migration_id = ? AND zabbix_server_id = ?", migration, serverId)
	return zabbixProxies, result.Error
}

// Store implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) Store(zabbixProxy *model.ZabbixProxy) error {
	result := r.db.Create(&zabbixProxy)
	return result.Error
}

// MultipleStore implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) MultipleStore(zabbixProxies []*model.ZabbixProxy) error {
	if len(zabbixProxies) == 0 {
		return nil
	}
	result := r.db.Create(&zabbixProxies)
	return result.Error
}

func (r *ZabbixProxyRepository) StoreMapping(mapping *model.ZabbixProxyMapping) error {
	result := r.db.Create(&mapping)
	return result.Error
}

func (r *ZabbixProxyRepository) Update(proxy *model.ZabbixProxy) error {
	result := r.db.Save(proxy)
	return result.Error
}
