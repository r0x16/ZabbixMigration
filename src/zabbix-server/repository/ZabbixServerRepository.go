package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/repository"
	"gorm.io/gorm"
)

type ZabbixServerRepository struct {
	db *gorm.DB
}

var _ repository.ZabbixServerRepository = &ZabbixServerRepository{}

func NewZabbixServerRepository(db *gorm.DB) *ZabbixServerRepository {
	repo := &ZabbixServerRepository{db: db}
	repo.db.AutoMigrate(&model.ZabbixServer{})
	return repo
}

// Store implements repository.ZabbixServerRepository.
func (r *ZabbixServerRepository) Store(zabbixServer *model.ZabbixServer) error {
	result := r.db.Create(&zabbixServer)
	return result.Error
}

// GetAll implements repository.ZabbixServerRepository.
func (r *ZabbixServerRepository) GetAll() ([]*model.ZabbixServer, error) {
	var zabbixServers []*model.ZabbixServer
	result := r.db.Find(&zabbixServers)
	return zabbixServers, result.Error
}

// GetByID implements repository.ZabbixServerRepository.
func (r *ZabbixServerRepository) GetByID(id uint) (*model.ZabbixServer, error) {
	var zabbixServer model.ZabbixServer
	result := r.db.First(&zabbixServer, id)
	return &zabbixServer, result.Error
}
