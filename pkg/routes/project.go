package routes

import (
	"fmt"
	"net/http"

	"github.com/3ssalunke/replc/pkg/controller"
	"github.com/labstack/echo/v4"
)

type (
	project struct {
		controller.Controller
	}

	projectRequest struct {
		Language string `json:"language" validate:"required"`
		ReplId   string `json:"replId" validate:"required"`
	}

	projectResponse struct {
		Msg   string `json:"msg"`
		Error string `json:"error"`
	}
)

func (c *project) Post(ctx echo.Context) error {
	payload := new(projectRequest)
	if err := ctx.Bind(payload); err != nil {
		return ctx.JSON(http.StatusBadRequest, &projectResponse{Error: err.Error()})
	}
	if err := c.Container.Validator.Validate(payload); err != nil {
		return ctx.JSON(http.StatusBadRequest, &projectResponse{Error: err.Error()})
	}
	storage, err := controller.NewS3Storage(*c.Container.Config)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, &projectResponse{Error: "Failed to spawn up a new replc"})
	}
	err = c.CopyObjects(payload.Language, payload.ReplId, storage)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, &projectResponse{Error: "Failed to spawn up a new replc"})
	}
	// return c.Redirect(ctx, "coding", fmt.Sprintf("?replcid=%s", payload.ReplId))
	url := fmt.Sprintf("/coding?replcid=%s", payload.ReplId)
	ctx.Response().Header().Set("HX-Redirect", url)
	return ctx.Redirect(http.StatusSeeOther, url)
}
