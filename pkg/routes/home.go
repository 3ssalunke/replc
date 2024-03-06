package routes

import (
	"github.com/3ssalunke/replc/pkg/controller"
	"github.com/3ssalunke/replc/templates"
	"github.com/labstack/echo/v4"
)

type (
	home struct {
		controller.Controller
	}
)

func (c *home) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageHome
	page.Metatags.Description = "Welcome to the replc."

	return c.RenderPage(ctx, page)
}
