package domain

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type EventClient interface {
	GetId() string
	Setup() *model.Error
	SendMessage(message *EventMessage) *model.Error
	Online() *model.Error
}
