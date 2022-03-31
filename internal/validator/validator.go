package validator

import (
	"encoding/json"
	"io/ioutil"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/labstack/echo/v4"
	"github.com/perses/poc-cuelang/internal/config"
	"github.com/perses/poc-cuelang/internal/model"
	"github.com/sirupsen/logrus"
)

type Validator interface {
	Validate(c echo.Context) error
	LoadSchemas(path string)
}

type validator struct {
	context *cue.Context
	schemas []cue.Value
}

func New(c *config.Config) Validator {
	// create a Context
	ctx := cuecontext.New()

	// initial load of schemas
	schemas, err := parseSchemas(ctx, c.SchemasPath)
	if err != nil {
		logrus.WithError(err).Error("not able to retrieve the list of schema files")
	}

	validator := &validator{
		context: ctx,
		schemas: schemas,
	}

	return validator
}

/*
 * Validate the received input.
 * The payload is matched against the known list of CUE definitions (schemas).
 * If no schema matches, the validation fails.
 */
func (v *validator) Validate(c echo.Context) error {
	// deserialize input into a Dashboard struct
	dashboard := new(model.Dashboard)
	err := c.Bind(dashboard)
	if err != nil {
		logrus.WithError(err).Error("Failed unmarshalling the received payload")
		return err
	}
	logrus.Tracef("Dashboard to validate: %+v", dashboard)

	var res error
	for _, panel := range dashboard.Spec.Panels {
		// remarshal the panel to be processed by CUE
		panelJson, _ := json.Marshal(panel)
		logrus.Tracef("Panel to validate: %s", string(panelJson))

		// compile the JSON panel into a CUE Value
		value := v.context.CompileBytes(panelJson)

		// iterate over schemas until we find a matching one for our value
		for _, schema := range v.schemas {
			logrus.Tracef("Matching panel against schema: %+v", schema)

			unified := value.Unify(schema)
			opts := []cue.Option{
				cue.Concrete(true),
				cue.Attributes(true),
				cue.Definitions(true),
				cue.Hidden(true),
			}

			err = unified.Validate(opts...)
			if err != nil {
				// validation error, but maybe the next schema will work
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
		logrus.Info("This dashboard is valid (all its panels are valid)")
	} else {
		logrus.WithError(err).Error("This dashboard is invalid (at least 1 panel is invalid)")
	}

	return res
}

/*
 * Load the known list of schemas into the validator
 */
func (v *validator) LoadSchemas(path string) {
	logrus.Info("Loading schemas")

	schemas, err := parseSchemas(v.context, path)
	if err != nil {
		logrus.WithError(err).Error("not able to retrieve the list of schema files")
		return
	}

	v.schemas = schemas
}

/*
 * Retrieve & parse schemas from .cue files
 */
func parseSchemas(context *cue.Context, path string) ([]cue.Value, error) {
	schemas := make([]cue.Value, 0)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return schemas, err
	}

	for _, file := range files {
		// Load Cue file into Cue build.Instances slice (the second arg is a configuration object, not used atm)
		buildInstance := load.Instances([]string{path + file.Name()}, nil)[0]
		// build Value from the Instance
		schemas = append(schemas, context.BuildInstance(buildInstance))
	}

	return schemas, nil
}
