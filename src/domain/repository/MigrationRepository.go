package repository

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
)

type MigrationRepository interface {
	Store(migration *model.Migration) error
	GetAll() ([]*model.Migration, error)
}
