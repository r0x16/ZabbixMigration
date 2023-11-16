package action

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/infraestructure/repository"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/zabbix"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var proxyMapMutex sync.Mutex

func SetupProxyMapping(c echo.Context, bundle *drivers.ApplicationBundle) error {

	migration, err := checkProxyImport(c, bundle)
	if err != nil {
		return echo.NewHTTPError(err.Code, err.Message)
	}

	var sourceProxies []*model.ZabbixProxy
	var destinationProxies []*model.ZabbixProxy
	if migration.IsProxyImported {
		sourceProxies, err = getImportedProxies(migration, &migration.Source, bundle)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		destinationProxies, err = getImportedProxies(migration, &migration.Destination, bundle)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	importEventsUrl := c.Echo().Reverse("ProxyMapFlow_importStatus", migration.ID)
	return c.Render(http.StatusOK, "migration/proxy-map", echo.Map{
		"title":              "Proxy mapping",
		"migration":          migration,
		"importEventsUrl":    importEventsUrl,
		"sourceProxies":      sourceProxies,
		"destinationProxies": destinationProxies,
	})
}

func getImportedProxies(migration *model.Migration, server *model.ZabbixServer, bundle *drivers.ApplicationBundle) ([]*model.ZabbixProxy, *model.Error) {
	repo := repository.NewZabbixProxyRepository(bundle.Database.Connection)
	proxies, err := repo.GetByMigrationAndServer(migration.ID, server.ID)

	if err != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return proxies, nil
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

	extractAndStoreProxies(bundle, migration, &migration.Source)
	extractAndStoreProxies(bundle, migration, &migration.Destination)

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

func extractAndStoreProxies(
	bundle *drivers.ApplicationBundle,
	migration *model.Migration,
	server *model.ZabbixServer,
) *model.Error {
	repo := repository.NewZabbixProxyRepository(bundle.Database.Connection)

	proxies, err := getProxiesFromApi(server)

	if err != nil {
		return err
	}

	setProxiesMigration(proxies, migration)
	setProxiesServer(proxies, server)

	storeError := repo.MultipleStore(proxies)

	if storeError != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: storeError.Error(),
		}
	}

	return nil
}

func getProxiesFromApi(server *model.ZabbixServer) ([]*model.ZabbixProxy, *model.Error) {
	api := zabbix.ServerConnector(server)
	api.Connect(server.Username, server.Password)

	proxies, err := api.Request(api.Body("proxy.get", model.ZabbixParams{
		"output":          "extend",
		"selectInterface": "extend",
		"selectHosts":     []string{"hostid"},
	}))

	if err != nil {
		return nil, err
	}

	proxyList, err := decodeProxies(proxies)

	if err != nil {
		return nil, err
	}

	setProxiesHostCount(proxyList)

	return proxyList, nil
}

func decodeProxies(proxies *model.ZabbixResponse) ([]*model.ZabbixProxy, *model.Error) {
	var proxyList []*model.ZabbixProxy
	modifiedJSON := strings.ReplaceAll(string(proxies.RawResult), `"interface":[]`, `"interface":null`)
	decodeError := json.Unmarshal([]byte(modifiedJSON), &proxyList)

	if decodeError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: decodeError.Error(),
		}
	}

	return proxyList, nil
}

func setProxiesHostCount(proxies []*model.ZabbixProxy) {
	for _, proxy := range proxies {
		proxy.HostCount = len(proxy.Hosts)
	}
}

func setProxiesMigration(proxies []*model.ZabbixProxy, migration *model.Migration) {
	for _, proxy := range proxies {
		proxy.MigrationID = migration.ID
	}
}

func setProxiesServer(proxies []*model.ZabbixProxy, server *model.ZabbixServer) {
	for _, proxy := range proxies {
		proxy.ZabbixServerID = server.ID
	}
}
