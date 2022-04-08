package validator

import (
	"errors"
	"fmt"
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
	schemas map[string]cue.Value
}

/*
 * Instanciate a validator
 */
func New(c *config.Config) Validator {
	ctx := cuecontext.New()

	schemas, err := loadSchemas(ctx, c.SchemasPath)
	if err != nil {
		logrus.WithError(err).Error("Not able to retrieve the list of schema files")
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

	// go through the panels list
	// the processing stops as soon as it detects an invalid panel TODO: can probably be improved
	for k, panel := range dashboard.Spec.Panels {
		logrus.Tracef("Panel to validate: %s", string(panel))

		// compile the JSON panel into a CUE Value
		value := v.context.CompileBytes(panel)

		// retrieve panel's kind
		kind, err := value.LookupPath(cue.ParsePath("kind")).String()
		if err != nil {
			logrus.Error("This panel doesn't contain the required \"kind\" field")
			res = err
			break
		}

		// retrieve the corresponding schema
		var schema cue.Value
		var ok bool
		if schema, ok = v.schemas[kind]; !ok {
			logrus.Errorf("%s is not valid panel: unknown kind %s", k, kind)
			res = errors.New("Unknown panel kind")
			break
		}
		logrus.Tracef("Matching panel %s against schema: %+v", k, schema)

		// do the validation
		unified := value.Unify(schema)
		opts := []cue.Option{
			cue.Concrete(true),
			cue.Attributes(true),
			cue.Definitions(true),
			cue.Hidden(true),
		}
		err = unified.Validate(opts...)
		if err != nil {
			logrus.Errorf("%s is not a valid %s panel", k, kind)
			res = err
			break
		}
	}

	if res == nil {
		logrus.Info("This dashboard is OK, all its panels are valid")
	} else {
		logrus.WithError(res).Error("This dashboard is KO, at least 1 panel is invalid")
	}

	return res
}

/*
 * Load the known list of schemas into the validator
 */
func (v *validator) LoadSchemas(path string) {
	schemas, err := loadSchemas(v.context, path)
	if err != nil {
		logrus.WithError(err).Error("Not able to retrieve the list of schema files")
		return
	}

	v.schemas = schemas
	logrus.Info("Schemas list (re)loaded")
}

/*
 * Load & return the known list of schemas
 */
func loadSchemas(context *cue.Context, path string) (map[string]cue.Value, error) {
	schemas := map[string]cue.Value{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return schemas, err
	}

	// process each .cue file to convert it into a CUE Value
	for _, file := range files {
		schemaPath := fmt.Sprintf("%s/%s", path, file.Name())

		// load Cue files into Cue build.Instances slice
		buildInstances := load.Instances([]string{}, &load.Config{Dir: schemaPath})
		// we strongly assume that only 1 buildInstance should be returned (corresponding to the #panel schema), otherwise we skip it
		// TODO can probably be improved
		if len(buildInstances) != 1 {
			logrus.Errorf("The number of build instances for %s is != 1, skipping this schema", schemaPath)
			continue
		}
		buildInstance := buildInstances[0]

		// check for errors on the instances (these are typically parsing errors)
		if buildInstance.Err != nil {
			logrus.WithError(buildInstance.Err).Errorf("Error retrieving schema for %s, skipping this schema", schemaPath)
			continue
		}

		// build Value from the Instance
		schema := context.BuildInstance(buildInstance)
		if schema.Err() != nil {
			logrus.WithError(schema.Err()).Errorf("Error during build for %s, skipping this schema", schemaPath)
			continue
		}

		// check if another schema for the same Kind was already registered
		kind, _ := schema.LookupPath(cue.ParsePath("kind")).String()
		if _, ok := schemas[kind]; ok {
			logrus.Errorf("Conflict caused by %s: a schema already exists for kind %s, skipping this schema", schemaPath, kind)
			continue
		}

		schemas[kind] = schema
		logrus.Debugf("Loaded schema %s from file %s: %+v", kind, schemaPath, schema)
	}

	return schemas, nil
}
