package events

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"github.com/labstack/echo/v4"
)

type EventClientEcho struct {
	Id           string
	eventChannel chan *domain.EventMessage
	context      echo.Context
}

var _ domain.EventClient = &EventClientEcho{}

// NewEventClientEcho creates a new EventClientEcho instance.
func NewEventClientEcho(id string, c echo.Context) *EventClientEcho {
	return &EventClientEcho{
		Id:           id,
		eventChannel: make(chan *domain.EventMessage),
		context:      c,
	}
}

// GetId implements domain.EventClient.
func (c EventClientEcho) GetId() string {
	return c.Id
}

// SendMessage implements domain.EventClient.
func (c EventClientEcho) SendMessage(message *domain.EventMessage) *model.Error {
	c.eventChannel <- message
	return nil
}

// WaitForMessage implements domain.EventClient.
func (c EventClientEcho) Online() *model.Error {
	for {
		select {
		case message := <-c.eventChannel:
			err := c.handleEvent(message)
			if err != nil {
				return err
			}
		case <-c.context.Request().Context().Done():
			return nil
		}
	}
}

func (c EventClientEcho) Close() {
	close(c.eventChannel)
}

func (c EventClientEcho) handleEvent(message *domain.EventMessage) *model.Error {
	data, err := json.Marshal(message.Data)
	if err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error processing event data",
			Data:    err,
		}
	}

	return c.transportEvent(message.Event, string(data))

}

func (c EventClientEcho) transportEvent(event string, data string) *model.Error {
	const format = "event:%s\ndata:%s\n\n"
	_, err := c.context.Response().Write([]byte(fmt.Sprintf(format, event, data)))
	if err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error sending event",
			Data:    err,
		}
	}

	c.context.Response().Flush()
	return nil
}
