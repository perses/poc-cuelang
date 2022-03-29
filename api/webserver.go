package api

import (
	"errors"
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
	schemasPath = "dev/schemas/"
)

type ServerAPI struct {
	utilsEcho.Register
	ctx     *cue.Context
	schemas []cue.Value
}

func NewServerAPI() *ServerAPI {
	// create a Context
	ctx := cuecontext.New()

	// retrieve the list of schema files
	files, err := ioutil.ReadDir(schemasPath)
	if err != nil {
		log.Fatal(err)
	}

	schemas := make([]cue.Value, 0)
	for _, file := range files {
		// Load Cue file into Cue build.Instances slice (the second arg is a configuration object, not used atm)
		buildInstance := load.Instances([]string{schemasPath + file.Name()}, nil)[0]
		// build Value from the Instance
		schemas = append(schemas, ctx.BuildInstance(buildInstance))
	}

	return &ServerAPI{
		ctx:     ctx,
		schemas: schemas,
	}
}

func (s *ServerAPI) RegisterRoute(e *echo.Echo) {
	e.POST("/validate", func(c echo.Context) error {
		data, err := ioutil.ReadAll(c.Request().Body)

		fmt.Println("User input :")
		fmt.Println(string(data))

		// compile the CUE data into a Value
		v := s.ctx.CompileBytes(data)

		// iterate over schemas until we find a matching one for our value
		res := errors.New("this input didn't match any known schemas")
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
			} else {
				fmt.Println("This panel definition is valid !")
				res = nil
				break
			}
		}

		return res
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
