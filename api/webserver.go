package api

import (
	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
	"github.com/perses/poc-cuelang/internal/validator"
)

type ServerAPI struct {
	utilsEcho.Register
	validator validator.Validator
}

func NewServerAPI(v validator.Validator) *ServerAPI {
	return &ServerAPI{
		validator: v,
	}
}

func (s *ServerAPI) RegisterRoute(e *echo.Echo) {
	e.POST("/validate", s.validator.Validate)
}

func (s *ServerAPI) Close() error {
	return nil
}
