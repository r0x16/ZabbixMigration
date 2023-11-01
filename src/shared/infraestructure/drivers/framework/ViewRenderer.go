package framework

import (
	"html/template"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/labstack/echo/v4"
)

func SetupRenderer(e *echo.Echo) {
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
		"reverse": e.Reverse,
	}
}

func SetupStatic(e *echo.Echo) {
	e.Static("/", "src/assets")
}
