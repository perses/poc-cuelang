package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
	model "github.com/perses/poc-cuelang/model"
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
		// deserialize input into a Dashboard struct
		dashboard := new(model.Dashboard)
		err := c.Bind(dashboard)
		if err != nil {
			fmt.Printf("Failed unmarshalling the received payload: %s\n", err)
			return err
		}
		fmt.Println("Dashboard to validate :")
		fmt.Printf("%+v\n", dashboard)

		var res error
		for _, panel := range dashboard.Spec.Panels {
			fmt.Println("Panel to validate :")
			fmt.Printf("%+v\n", panel)

			// remarshal the panel to be processed by CUE
			panelJson, _ := json.Marshal(panel)
			fmt.Printf("After remarshal : %s\n", string(panelJson))

			// compile the JSON panel into a CUE Value
			v := s.ctx.CompileBytes(panelJson)

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
					res = nil
					break
				}
			}

			// an invalid panel was found, stop the processing here
			if res != nil {
				break
			}
		}

		if res == nil {
			fmt.Println("This dashboard is valid !")
		} else {
			fmt.Printf("This dashboard is not valid, at least 1 of its panels is invalid: %s\n", err)
		}

		return res
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
