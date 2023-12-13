package repository

import (
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/repository"
	"gorm.io/gorm"
)

type MigrationRepository struct {
	db    *gorm.DB
	mutex sync.Mutex
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
	result := r.db.Joins("Source").Joins("Destination").Order("migrations.id desc").Find(&migrations)
	return migrations, result.Error
}

// Store implements repository.MigrationRepository.
func (r *MigrationRepository) Store(migration *model.Migration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	result := r.db.Create(&migration)
	return result.Error
}

// GetById implements repository.MigrationRepository.
func (r *MigrationRepository) GetById(id uint) (*model.Migration, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var migration model.Migration
	result := r.db.Joins("Source").Joins("Destination").First(&migration, id)
	return &migration, result.Error
}

// Update implements repository.MigrationRepository.
func (r *MigrationRepository) Update(migration *model.Migration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	result := r.db.Save(&migration)
	return result.Error
}
