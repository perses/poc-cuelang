package watcher

import (
	"github.com/perses/poc-cuelang/internal/config"
	"github.com/perses/poc-cuelang/internal/validator"
	"github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

/*
 * Start watching for changes in the schemas folder.
 * Whenever a change is detected (write, creat, delete..), we
 * notify the validator to refresh its list of schemas.
 */
func Start(c *config.Config, v validator.Validator) {
	logrus.Debugf("Start watching file: %s", c.SchemasPath)

	go func() {
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
					if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove {
						logrus.Tracef("%s event on %s", event.Op, event.Name)
						v.LoadSchemas(c.SchemasPath)
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					logrus.WithError(err).Trace("watcher error")
				}
			}
		}()

		err = watcher.Add(c.SchemasPath)
		if err != nil {
			logrus.Fatal(err)
		}
		<-done
	}()
}
