package routes

import (
	"os"
	"path/filepath"

	"github.com/3ssalunke/replc/pkg/controller"
	"github.com/3ssalunke/replc/templates"
	"github.com/labstack/echo/v4"
)

type (
	coding struct {
		controller.Controller
	}
)

func (c *coding) Get(ctx echo.Context) error {
	replId := ctx.Param("replcid")

	k8s, err := controller.NewK8S(c.Container.Config.K8S.Kubeconfigpath)
	if err != nil {
		return c.Fail(err, "failed to load kube configs")
	}

	// Get the current directory
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return c.Fail(err, "failed to get current directory")
	}

	orchestratorYamlPath := filepath.Join(dir, "../orchestrator/service.yaml")

	err = c.Controller.CreateK8sResources(k8s, orchestratorYamlPath, replId)
	if err != nil {
		return c.Fail(err, "failed to load kube configs")
	}

	page := controller.NewPage(ctx)
	page.Layout = templates.LayoutMain
	page.Name = templates.PageCoding
	page.Metatags.Description = "Welcome to the replc."
	page.Data = struct{ ReplId string }{
		ReplId: replId,
	}

	return c.RenderPage(ctx, page)
}
