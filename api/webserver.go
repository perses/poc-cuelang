package api

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sirupsen/logrus"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/labstack/echo/v4"
	utilsEcho "github.com/perses/common/echo"
	config "github.com/perses/poc-cuelang/internal/config"
	model "github.com/perses/poc-cuelang/internal/model"
)

type ServerAPI struct {
	utilsEcho.Register
	ctx     *cue.Context
	schemas []cue.Value
}

func NewServerAPI(c *config.Config) *ServerAPI {
	// create a Context
	ctx := cuecontext.New()

	// retrieve the list of schema files
	files, err := ioutil.ReadDir(c.SchemasPath)
	if err != nil {
		logrus.WithError(err).Fatal("not able to retrieve the list of schema files")
	}

	schemas := make([]cue.Value, 0)
	for _, file := range files {
		// Load Cue file into Cue build.Instances slice (the second arg is a configuration object, not used atm)
		buildInstance := load.Instances([]string{c.SchemasPath + file.Name()}, nil)[0]
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
			logrus.WithError(err).Error("Failed unmarshalling the received payload")
			return err
		}
		logrus.Tracef("Dashboard to validate : %+v", dashboard)

		var res error
		for _, panel := range dashboard.Spec.Panels {
			// remarshal the panel to be processed by CUE
			panelJson, _ := json.Marshal(panel)
			logrus.Tracef("Panel to validate : %s", string(panelJson))

			// compile the JSON panel into a CUE Value
			v := s.ctx.CompileBytes(panelJson)

			// iterate over schemas until we find a matching one for our value
			for _, schema := range s.schemas {
				logrus.Tracef("Matching panel against schema : %+v", schema)

				unified := v.Unify(schema)
				opts := []cue.Option{
					cue.Concrete(true),
					cue.Attributes(true),
					cue.Definitions(true),
					cue.Hidden(true),
				}

				err = unified.Validate(opts...)
				if err != nil {
					// Validation error, but maybe the next schema will work
					res = err
				} else {
					logrus.Debug("This panel is valid (found matching schema)")
					res = nil
					break
				}
			}

			// an invalid panel was found, stop the processing here
			if res != nil {
				logrus.WithError(err).Error("This panel is invalid, no schema corresponds")
				break
			}
		}

		if res == nil {
			logrus.Info("This dashboard is valid (all its panels are valid")
		} else {
			logrus.WithError(err).Error("This dashboard is invalid (at least 1 panel is invalid)")
		}

		return res
	})
}

func (s *ServerAPI) Close() error {
	return nil
}
