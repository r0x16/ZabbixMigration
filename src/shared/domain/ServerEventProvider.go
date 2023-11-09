package domain

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ServerEventProvider interface {
	Subscribe(client EventClient) *model.Error
	Unsubscribe(client EventClient) *model.Error
	Broadcast(message *EventMessage) *model.Error
}
