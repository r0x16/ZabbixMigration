package framework

import (
	"html/template"
	"time"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure/drivers/ui"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/labstack/echo/v4"
)

func SetupRenderer(e *echo.Echo) {
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		panic(err)
	}
	time.Local = loc

	e.Renderer = echoview.New(goview.Config{
		Root:         "src/views",
		Extension:    ".html",
		Master:       "layouts/master",
		Partials:     []string{},
		Funcs:        viewFuncs(e),
		DisableCache: true,
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	})
}

func viewFuncs(e *echo.Echo) template.FuncMap {
	return template.FuncMap{
		"reverse":     e.Reverse,
		"date_format": ui.DateFormat,
	}
}

func SetupStatic(e *echo.Echo) {
	e.Static("/", "src/assets")
}
