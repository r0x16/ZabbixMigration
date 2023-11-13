package action

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var proxyMapMutex sync.Mutex

func SetupProxyMapping(c echo.Context, bundle *drivers.ApplicationBundle) error {

	migration, err := checkProxyImport(c, bundle)
	if err != nil {
		return echo.NewHTTPError(err.Code, err.Message)
	}

	importEventsUrl := c.Echo().Reverse("ProxyMapFlow_importStatus", migration.ID)
	return c.Render(http.StatusOK, "migration/proxy-map", echo.Map{
		"title":           "Proxy mapping",
		"migration":       migration,
		"importEventsUrl": importEventsUrl,
	})
}

func ImportProxyStatusEvents(c echo.Context, bundle *drivers.ApplicationBundle) error {
	client := events.NewEventClientEcho(uuid.New().String(), c)
	err := client.Setup()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	migration, err := checkProxyImport(c, bundle)

	if err != nil {
		return echo.NewHTTPError(err.Code, err.Message)
	}

	if migration.IsProxyImported {
		client.Close()
		return c.NoContent(http.StatusOK)
	}

	event_id := fmt.Sprintf("proxy-import-%d", migration.ID)
	bundle.ServerEvents[event_id].Subscribe(client)

	client.Online()
	eventHandler, active := bundle.ServerEvents[event_id]
	if active {
		eventHandler.Unsubscribe(client)
	}
	return nil
}

func checkProxyImport(c echo.Context, bundle *drivers.ApplicationBundle) (*model.Migration, *model.Error) {
	proxyMapMutex.Lock()
	defer proxyMapMutex.Unlock()

	migration, err := GetMigrationFromParam(c, bundle)
	if err != nil {
		return nil, err
	}

	if !migration.IsProxyImported {
		startProxyImport(bundle, migration)
	}

	return migration, nil
}

func startProxyImport(bundle *drivers.ApplicationBundle, migration *model.Migration) {
	event_id := fmt.Sprintf("proxy-import-%d", migration.ID)
	_, started := bundle.ServerEvents[event_id]
	if !started {
		bundle.ServerEvents[event_id] = events.NewServerEventEcho(event_id)
		go importProxies(bundle, migration)
	}
}

func importProxies(bundle *drivers.ApplicationBundle, migration *model.Migration) {
	event_id := fmt.Sprintf("proxy-import-%d", migration.ID)
	eventHandler := bundle.ServerEvents[event_id]
	// TO-DO: implement proxy import
	f, err := os.Create("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer f.Close()

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		f.WriteString(fmt.Sprintf("seconds elapsed: %s\n", strconv.Itoa(i)))
	}

	closeError := closeImportProxy(bundle, migration, eventHandler)

	if closeError != nil {
		eventHandler.Broadcast(&domain.EventMessage{
			Event: "error",
			Data:  closeError,
		})
	}

	delete(bundle.ServerEvents, event_id)
}

func closeImportProxy(bundle *drivers.ApplicationBundle, migration *model.Migration, eventHandler domain.ServerEventProvider) *model.Error {
	proxyMapMutex.Lock()
	defer proxyMapMutex.Unlock()

	migration.IsProxyImported = true
	repo := repository.NewMigrationRepository(bundle.Database.Connection)
	if err := repo.Update(migration); err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	eventHandler.Broadcast(&domain.EventMessage{
		Event: "ready",
		Data:  "Import finished",
	})

	return nil
}
