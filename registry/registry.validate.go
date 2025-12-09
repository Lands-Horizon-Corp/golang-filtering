package registry

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Registry[TData, TResponse, TRequest]) Validate(ctx echo.Context) (*TRequest, error) {
	var req TRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := r.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}
