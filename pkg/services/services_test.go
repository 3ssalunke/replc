package services

import (
	"os"
	"testing"

	"github.com/3ssalunke/replc/config"
	"github.com/3ssalunke/replc/pkg/tests"
	"github.com/labstack/echo/v4"
)

var (
	c   *Container
	ctx echo.Context
)

func TestMain(m *testing.M) {
	// Set the environment to test
	config.SwitchEnvironment(config.EnvTest)

	// Create a new container
	c = NewContainer()

	// Create a web context
	ctx, _ = tests.NewContext(c.Web, "/")
	tests.InitSession(ctx)

	// Run tests
	exitVal := m.Run()

	// Shutdown the container
	if err := c.Shutdown(); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
