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
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/events"
	"github.com/labstack/echo/v4"
)

var setupProxyMutex sync.Mutex

func SetupProxyMapping(c echo.Context, bundle *drivers.ApplicationBundle) error {

	imported, err := checkProxyImport(c, bundle)
	if err != nil {
		return echo.NewHTTPError(err.Code, err.Message)
	}

	return c.Render(http.StatusOK, "migration/proxy-map", echo.Map{
		"title":    "Proxy mapping",
		"imported": imported,
	})
}

func checkProxyImport(c echo.Context, bundle *drivers.ApplicationBundle) (bool, *model.Error) {
	setupProxyMutex.Lock()
	defer setupProxyMutex.Unlock()

	migration, err := GetMigrationFromParam(c, bundle)
	if err != nil {
		return false, err
	}

	if !migration.IsProxyImported {
		startProxyImport(bundle, migration)
		return false, nil
	}

	return true, nil
}

func startProxyImport(bundle *drivers.ApplicationBundle, migration *model.Migration) {
	event_id := fmt.Sprintf("proxy-import-%d", migration.ID)
	event, started := bundle.ServerEvents[event_id]
	if !started {
		bundle.ServerEvents[event_id] = events.NewServerEventEcho(event_id)
		go importProxies(bundle, migration)
	}

	log.Println(event)
}

func importProxies(bundle *drivers.ApplicationBundle, migration *model.Migration) {
	// TO-DO: implement proxy import
	f, err := os.Create("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file
	defer f.Close()

	for i := 0; i < 60; i++ {
		time.Sleep(1 * time.Second)
		f.WriteString(fmt.Sprintf("seconds elapsed: %s\n", strconv.Itoa(i)))
	}
}
