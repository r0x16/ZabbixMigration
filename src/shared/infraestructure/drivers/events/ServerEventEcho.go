package events

import (
	"net/http"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
)

type ServerEventEcho struct {
	EventId string
	Clients map[string]EventClientEcho
	m       sync.Mutex
}

var _ domain.ServerEventProvider = &ServerEventEcho{}

// NewServerEventEcho creates a new ServerEventEcho instance.
func NewServerEventEcho(id string) *ServerEventEcho {
	return &ServerEventEcho{
		EventId: id,
		Clients: make(map[string]EventClientEcho),
	}
}

// Subscribe implements domain.ServerEventProvider.
func (se *ServerEventEcho) Subscribe(client domain.EventClient) *model.Error {
	se.m.Lock()
	defer se.m.Unlock()

	id := client.GetId()

	if _, ok := se.Clients[id]; ok {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "client already subscribed",
		}
	}

	se.Clients[id] = client.(EventClientEcho)
	return nil
}

// Unsubscribe implements domain.ServerEventProvider.
func (se *ServerEventEcho) Unsubscribe(client domain.EventClient) *model.Error {
	se.m.Lock()
	defer se.m.Unlock()

	id := client.GetId()

	if _, ok := se.Clients[id]; !ok {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "client not subscribed",
		}
	}

	delete(se.Clients, id)
	return nil
}

// Broadcast implements domain.ServerEventProvider.
func (se *ServerEventEcho) Broadcast(message *domain.EventMessage) *model.Error {
	se.m.Lock()
	defer se.m.Unlock()

	for _, client := range se.Clients {
		err := client.SendMessage(message)
		if err != nil {
			return err
		}
	}

	return nil
}
