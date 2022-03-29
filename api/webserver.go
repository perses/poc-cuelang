package api

import (
	"fmt"
	"io/ioutil"
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
)

const (
	cueFile = "dev/schemas/line.cue"
)

type ServerAPI struct {
	utilsEcho.Register
	ctx     *cue.Context
	schemas []cue.Value
}

func NewServerAPI() *ServerAPI {
	// create a Context
	ctx := cuecontext.New()
	// Retrieve our schemas.
	entrypoints := []string{cueFile}
	// - Load Cue files into Cue build.Instances slice (the second arg is a configuration object, not used atm)
	buildInstances := load.Instances(entrypoints, nil)
	// - build Values from the Instances
	schemas, err := ctx.BuildInstances(buildInstances)
	// check for errors on the instances, these are typically parsing errors
	if err != nil {
		log.Fatalf("Error retrieving schemas : %v\n", err)
	}

	return &ServerAPI{
		ctx:     ctx,
		schemas: schemas,
	}
}

func (s *ServerAPI) RegisterRoute(e *echo.Echo) {
	e.POST("/validate", func(c echo.Context) error {
		var res error
		data, err := ioutil.ReadAll(c.Request().Body)

		fmt.Println("User input :")
		fmt.Println(string(data))

		// compile the CUE data into a Value
		v := s.ctx.CompileBytes(data)

		// iterate over schemas until we find a matching one for our value
		for _, schema := range s.schemas {
			fmt.Printf("Current schema : %v\n", schema)

			unified := v.Unify(schema)
			opts := []cue.Option{
				cue.Attributes(true),
				cue.Definitions(true),
				cue.Hidden(true),
			}

			err = unified.Validate(opts...)
			if err != nil {
				fmt.Printf("Validation Error: %s\n", err)
				res = err
			} else {
				fmt.Println("This panel definition is valid !")
			}
		}

		return res
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
