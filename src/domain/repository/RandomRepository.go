package repository

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type RandomRepository interface {
	GenerateRandomString(int) *model.Random
	Store(*model.Random) error
	GetAll() ([]*model.Random, error)
}