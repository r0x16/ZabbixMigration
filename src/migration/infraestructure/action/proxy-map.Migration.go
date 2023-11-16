package action

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	mdomain "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/migration/domain"
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

	var storeError *model.Error
	if c.Request().Method == http.MethodPost {
		storeError = storeProxyMapping(migration, c, bundle)
		if storeError == nil {
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("MigrationCreate", migration.ID))
		}
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
		"error":              storeError,
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

func storeProxyMapping(migration *model.Migration, c echo.Context, bundle *drivers.ApplicationBundle) *model.Error {

	mappings, bindingError := bindProxyMappings(c, migration, bundle)
	if bindingError != nil {
		return bindingError
	}

	validationError := validateProxyMappings(mappings)
	if validationError != nil {
		return validationError
	}

	storeError := storeMappings(mappings, bundle)
	if storeError != nil {
		return storeError
	}

	migration.DefaultProxyID = sql.NullInt32{Int32: int32(mappings.DefaultProxy), Valid: true}

	markError := updateMigrationAsProxyMapped(migration, bundle)
	if markError != nil {
		return markError
	}

	return nil
}

func updateMigrationAsProxyMapped(migration *model.Migration, bundle *drivers.ApplicationBundle) *model.Error {
	migration.IsProxyMapped = true
	repo := repository.NewMigrationRepository(bundle.Database.Connection)
	err := repo.Update(migration)

	if err != nil {
		return &model.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

func storeMappings(mappings *mdomain.ProxyMappingBody, bundle *drivers.ApplicationBundle) *model.Error {
	for index, sourceProxy := range mappings.SourceProxies {
		destinationProxy := mappings.DestinationProxies[index]
		proxyMap := &model.ZabbixProxyMapping{
			SourceProxyID:      sourceProxy,
			DestinationProxyID: destinationProxy,
		}

		repo := repository.NewZabbixProxyRepository(bundle.Database.Connection)
		err := repo.StoreMapping(proxyMap)

		if err != nil {
			return &model.Error{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
	}
	return nil
}

func validateProxyMappings(mappings *mdomain.ProxyMappingBody) *model.Error {

	requiredParametersError := validateProxyMappingsRequired(mappings)
	if requiredParametersError != nil {
		return requiredParametersError
	}

	srcProxy, proxyPresentError := validateProxyPresent(mappings.SourceProxies, mappings.ImportedSourceProxies)
	if !proxyPresentError {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Source proxy %d not found", srcProxy),
		}
	}

	dstProxy, proxyPresentError := validateProxyPresent(mappings.DestinationProxies, mappings.ImportedDestinationProxies)
	if !proxyPresentError {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Destination proxy %d not found", dstProxy),
		}
	}

	return nil
}

func validateProxyPresent(proxyIds []uint, proxies []*model.ZabbixProxy) (uint, bool) {
	for _, proxyId := range proxyIds {
		proxyPresent := false
		for _, proxy := range proxies {
			if proxy.ID == proxyId {
				proxyPresent = true
				break
			}
		}

		if !proxyPresent {
			return proxyId, false
		}
	}

	return 0, true
}

func validateProxyMappingsRequired(mappings *mdomain.ProxyMappingBody) *model.Error {
	if mappings.DefaultProxy == 0 {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Default proxy is required",
		}
	}

	if len(mappings.SourceProxies) != len(mappings.ImportedSourceProxies) {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Invalid source proxies",
		}
	}

	if len(mappings.SourceProxies) != len(mappings.DestinationProxies) {
		return &model.Error{
			Code:    http.StatusBadRequest,
			Message: "Source and destination proxies must be the same length",
		}
	}

	return nil
}

func bindProxyMappings(c echo.Context, migration *model.Migration, bundle *drivers.ApplicationBundle) (*mdomain.ProxyMappingBody, *model.Error) {
	var proxyMappingBody mdomain.ProxyMappingBody
	bindingError := c.Bind(&proxyMappingBody)
	if bindingError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: bindingError.Error(),
		}
	}

	var importError *model.Error
	proxyMappingBody.ImportedSourceProxies, importError = getImportedProxies(migration, &migration.Source, bundle)
	if importError != nil {
		return nil, importError
	}

	proxyMappingBody.ImportedDestinationProxies, importError = getImportedProxies(migration, &migration.Destination, bundle)
	if importError != nil {
		return nil, importError
	}

	return &proxyMappingBody, nil
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
