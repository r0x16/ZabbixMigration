package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/repository"
	"gorm.io/gorm"
)

type ZabbixHostRepository struct {
	db *gorm.DB
}

var _ repository.ZabbixHostRepository = &ZabbixHostRepository{}

func NewZabbixHostRepository(db *gorm.DB) *ZabbixHostRepository {
	repo := &ZabbixHostRepository{db: db}
	repo.db.AutoMigrate(&model.ZabbixHost{})
	return repo
}

// FindByMigration implements repository.ZabbixHostRepository.
func (r *ZabbixHostRepository) FindByMigration(migration *model.Migration) ([]*model.ZabbixHost, error) {
	var zabbixHosts []*model.ZabbixHost
	result := r.db.Find(&zabbixHosts, "migration_id = ?", migration.ID)
	return zabbixHosts, result.Error
}

// FindByMigrationAndProxy implements repository.ZabbixHostRepository.
func (r *ZabbixHostRepository) FindByMigrationAndProxy(migration *model.Migration, proxy *model.ZabbixProxy) ([]*model.ZabbixHost, error) {
	var zabbixHosts []*model.ZabbixHost
	result := r.db.Find(&zabbixHosts, "migration_id = ? AND proxy_host_id = ?", migration.ID, proxy.ProxyID)
	return zabbixHosts, result.Error
}

// MultipleStore implements repository.ZabbixHostRepository.
func (r *ZabbixHostRepository) MultipleStore(hosts []*model.ZabbixHost) error {
	if len(hosts) == 0 {
		return nil
	}

	result := r.db.Create(&hosts)
	return result.Error
}

// Update implements repository.ZabbixHostRepository.
func (r *ZabbixHostRepository) Update(host *model.ZabbixHost) error {
	result := r.db.Save(host)
	return result.Error
}
