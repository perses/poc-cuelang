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
	"gopkg.in/fsnotify.v1"
)

type Validator interface {
	Validate(c echo.Context) error
}

type validator struct {
	ctx     *cue.Context
	schemas []cue.Value
}

func New(c *config.Config) *validator {
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

	go watchSchemas(c.SchemasPath)

	return &validator{
		ctx:     ctx,
		schemas: schemas,
	}
}

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
		value := v.ctx.CompileBytes(panelJson)

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
		logrus.Info("This dashboard is valid (all its panels are valid)")
	} else {
		logrus.WithError(err).Error("This dashboard is invalid (at least 1 panel is invalid)")
	}

	return res
}

// watch schemas for changes
// TODO useless watch for now, just prints events.
// TODO to be moved to a dedicated package
func watchSchemas(filepath string) {
	logrus.Debugf("Start watching file: %s", filepath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					logrus.Tracef("%s created", event.Name)
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					logrus.Tracef("%s modified", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					logrus.Tracef("%s moved or deleted", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.WithError(err).Trace("watcher error")
			}
		}
	}()

	err = watcher.Add(filepath)
	if err != nil {
		logrus.Fatal(err)
	}
	<-done
}
