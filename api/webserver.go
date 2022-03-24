package api

import (
	"fmt"
	"io/ioutil"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
	persesModel "github.com/perses/perses/pkg/model/api/v1"
)

const (
	schemaName = "datasourceSchema"
	schemaDef  = "#" + schemaName + `: {
		kind: "Datasource"
		metadata: {
			name: string
		}
		spec: {
			kind: string | *"Prometheus"
			default: bool
			http: {
				url: string
			}
		}
	}
	`
)

type ServerAPI struct {
	utilsEcho.Register
	ctx    *cue.Context
	schema cue.Value
}

func NewServerAPI() *ServerAPI {
	// create a Context
	ctx := cuecontext.New()
	// compile our schema
	schema := ctx.CompileString(schemaDef)

	return &ServerAPI{
		ctx:    ctx,
		schema: schema,
	}
}

func (s *ServerAPI) RegisterRoute(e *echo.Echo) {
	e.POST("/validate", func(c echo.Context) error {
		data, err := ioutil.ReadAll(c.Request().Body)

		fmt.Println("User input :")
		fmt.Println(string(data))

		// build the final CUE payload as a combination of user input & schema constraint
		cueData := fmt.Sprintf("#%s & %s", schemaName, data)

		// compile the CUE data into a Value, with scope
		v := s.ctx.CompileString(cueData, cue.Scope(s.schema))

		// check for errors during compiling
		if v.Err() != nil {
			msg := errors.Details(v.Err(), nil)
			fmt.Printf("Compile Error:\n%s\n", msg)
		} else {
			fmt.Println("CUE compilation result :")
			fmt.Println(v)
		}

		// evaluate the CUE Value
		e := v.Eval()
		if e.Err() != nil {
			msg := errors.Details(e.Err(), nil)
			fmt.Printf("Eval Error:\n%s\n", msg)
		} else {
			fmt.Println("CUE evaluation result :")
			fmt.Printf("%#v\n", e)
		}

		// check if the CUE Value is concrete
		if v.IsConcrete() {
			fmt.Println("CUE Value is concrete")
		} else {
			fmt.Println("CUE Value is not concrete")
		}

		// decode the CUE Value into Go struct
		var datasource persesModel.Datasource
		decodeErr := v.Decode(&datasource)
		if decodeErr != nil {
			fmt.Printf("Decode Error:\n%s\n", decodeErr)
		} else {
			fmt.Println("CUE Value successfully decoded into a Datasource object :")
			fmt.Println(datasource)
		}

		return err
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
