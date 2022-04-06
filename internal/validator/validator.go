package validator

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/labstack/echo/v4"
	"github.com/perses/poc-cuelang/internal/config"
	"github.com/perses/poc-cuelang/internal/model"
	"github.com/sirupsen/logrus"
)

const baseDefPath = "cue/base.cue"

type Validator interface {
	Validate(c echo.Context) error
	LoadSchemas(path string)
}

type validator struct {
	context *cue.Context
	baseDef cue.Value
	schemas map[string]cue.Value
}

/*
 * Instanciate a validator
 */
func New(c *config.Config) Validator {
	ctx := cuecontext.New()

	// load the base panel definition
	data, err := os.ReadFile(baseDefPath)
	if err != nil {
		logrus.WithError(err).Fatalf("Not able to read the base panel definition file %s, shutting down..", baseDefPath)
	}
	baseDef := ctx.CompileBytes(data)

	schemas, err := loadSchemas(ctx, baseDef, c.SchemasPath)
	if err != nil {
		logrus.WithError(err).Error("Not able to retrieve the list of schema files")
	}

	validator := &validator{
		context: ctx,
		baseDef: baseDef,
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
	for _, panel := range dashboard.Spec.Panels {
		logrus.Tracef("Panel to validate: %s", string(panel))

		// compile the JSON panel into a CUE Value
		value := v.context.CompileBytes(panel)

		// retrieve panel's kind
		kind, err := value.LookupPath(cue.ParsePath("kind")).String()
		if err != nil {
			logrus.WithError(err).Error("This panel doesn't embed the required Kind property")
			break
		}

		// retrieve the corresponding schema
		var schema cue.Value
		var ok bool
		if schema, ok = v.schemas[kind]; !ok {
			notFound := errors.New("Unknown panel kind")
			logrus.Errorf("This panel is not valid: unknown kind %s", kind)
			res = notFound
			break
		}
		logrus.Tracef("Matching panel against schema: %+v", schema)

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
			logrus.WithError(err).Errorf("This panel is not a valid %s", kind)
			res = err
			break
		}
	}

	if res == nil {
		logrus.Info("This dashboard is OK, all its panels are valid")
	} else {
		logrus.WithError(err).Error("This dashboard is KO, at least 1 panel is invalid")
	}

	return res
}

/*
 * Load the known list of schemas into the validator
 */
func (v *validator) LoadSchemas(path string) {
	schemas, err := loadSchemas(v.context, v.baseDef, path)
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
func loadSchemas(context *cue.Context, baseDef cue.Value, path string) (map[string]cue.Value, error) {
	schemas := map[string]cue.Value{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return schemas, err
	}

	// process each .cue file to convert it into a CUE Value
	// for each schema we check that it meets the default specs we expect for any panel, otherwise we dont include it
	for _, file := range files {
		schemaPath := fmt.Sprintf("%s/%s", path, file.Name())
		entrypoints := []string{schemaPath, baseDefPath}

		// load Cue files into Cue build.Instances slice (the second arg is a configuration object that we dont need here)
		buildInstances := load.Instances(entrypoints, nil)
		// we strongly assume that only 1 buildInstance should be returned (corresponding to the #panel schema), otherwise we skip it
		// TODO can probably be improved
		if len(buildInstances) != 1 {
			logrus.Errorf("The number of build instances for %s is != 1, skipping this schema", file.Name())
			continue
		}
		buildInstance := buildInstances[0]

		// check for errors on the instances (these are typically parsing errors)
		if buildInstance.Err != nil {
			logrus.WithError(buildInstance.Err).Errorf("Error retrieving schema for %s, skipping this schema", file.Name())
			continue
		}

		// build Value from the Instance
		schema := context.BuildInstance(buildInstance)
		if schema.Err() != nil {
			logrus.WithError(schema.Err()).Errorf("Error during build for %s, skipping this schema", file.Name())
			continue
		}

		// check if another schema for the same Kind was already registered
		kind, _ := schema.LookupPath(cue.ParsePath("kind")).String()
		if _, ok := schemas[kind]; ok {
			logrus.Errorf("Conflict caused by %s: a schema already exists for kind %s, skipping this schema", file.Name(), kind)
			continue
		}

		schemas[kind] = schema
		logrus.Tracef("Loaded new schema from file %s: %+v", file.Name(), schema)
	}

	return schemas, nil
}
