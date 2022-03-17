package api

import (
	"io/ioutil"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	utilsEcho "github.com/perses/common/echo"
	log "github.com/sirupsen/logrus"
)

func timestampNow() string {
	return strconv.Itoa(int(time.Now().UnixNano()))
}

type ServerAPI struct {
	utilsEcho.Register
	data *cache.Cache
}

func NewServerAPI(expiration, cleanupInterval *time.Duration) *ServerAPI {
	return &ServerAPI{}
}

func (s *ServerAPI) RegisterRoute(e *echo.Echo) {
	e.POST("/validate", func(c echo.Context) error {
		var err error
		s.data, err = ioutil.ReadAll(c.Request().Body)
		log.Infof("New payload to validate : %s", s.data)
		return err
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
