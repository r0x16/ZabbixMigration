package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
)

type MigrationRepository interface {
	Store(migration *model.Migration) error
	GetAll() ([]*model.Migration, error)
	GetById(id uint) (*model.Migration, error)
	Update(migration *model.Migration) error
}
