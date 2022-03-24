package api

import (
	"fmt"
	"io/ioutil"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
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

		// create a Context
		ctx := cuecontext.New()

		// compile the input into a Value  (= cue evaluation)
		var v cue.Value = ctx.CompileBytes(data)

		// print the value
		fmt.Println(v)

		return err
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
