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
	return repo
}

// GetAll implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) GetAll() ([]*model.ZabbixProxy, error) {
	var zabbixProxies []*model.ZabbixProxy
	result := r.db.Joins("Interface").Find(&zabbixProxies)
	return zabbixProxies, result.Error
}

// Store implements repository.ZabbixProxyRepository.
func (r *ZabbixProxyRepository) Store(zabbixProxy *model.ZabbixProxy) error {
	result := r.db.Create(&zabbixProxy)
	return result.Error
}
