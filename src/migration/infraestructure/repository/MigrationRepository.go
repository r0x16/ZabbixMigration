package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/repository"
	"gorm.io/gorm"
)

type MigrationRepository struct {
	db *gorm.DB
}

var _ repository.MigrationRepository = &MigrationRepository{}

func NewMigrationRepository(db *gorm.DB) *MigrationRepository {
	repo := &MigrationRepository{db: db}
	repo.db.AutoMigrate(&model.Migration{})
	return repo
}

// GetAll implements repository.MigrationRepository.
func (r *MigrationRepository) GetAll() ([]*model.Migration, error) {
	var migrations []*model.Migration
	result := r.db.Joins("Source").Joins("Destination").Find(&migrations)
	return migrations, result.Error
}

// Store implements repository.MigrationRepository.
func (r *MigrationRepository) Store(migration *model.Migration) error {
	result := r.db.Create(&migration)
	return result.Error
}
