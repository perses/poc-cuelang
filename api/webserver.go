package api

import (
	"io/ioutil"

	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
	log "github.com/sirupsen/logrus"
)

type ServerAPI struct {
	utilsEcho.Register
}

func NewServerAPI() *ServerAPI {
	return &ServerAPI{}
}

func (s *ServerAPI) RegisterRoute(e *echo.Echo) {
	e.POST("/validate", func(c echo.Context) error {
		data, err := ioutil.ReadAll(c.Request().Body)
		log.Infof("New payload to validate : %s", data)
		return err
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
