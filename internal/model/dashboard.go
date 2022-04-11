// fork from main Perses project, as we need the dashboard struct without the panel struct embedded

package model

import (
	"encoding/json"

	perses "github.com/perses/perses/pkg/model/api/v1"
	"github.com/perses/perses/pkg/model/api/v1/common"
	"github.com/perses/perses/pkg/model/api/v1/dashboard"
	"github.com/prometheus/common/model"
)

type Dashboard struct {
	Kind     perses.Kind            `json:"kind" yaml:"kind"`
	Metadata perses.ProjectMetadata `json:"metadata" yaml:"metadata"`
	Spec     DashboardSpec          `json:"spec" yaml:"spec"`
}

type DashboardSpec struct {
	// Datasource is a set of values that will be used to find the datasource definition.
	Datasource dashboard.Datasource `json:"datasource" yaml:"datasource"`
	// Duration is the default time you would like to use to looking in the past when getting data to fill the
	// dashboard
	Duration   model.Duration                 `json:"duration" yaml:"duration"`
	Variables  map[string]*dashboard.Variable `json:"variables,omitempty" yaml:"variables,omitempty"`
	Panels     map[string]json.RawMessage     `json:"panels" yaml:"panels"` // Part modified from Perses datamodel
	Layouts    map[string]*dashboard.Layout   `json:"layouts" yaml:"layouts"`
	Entrypoint *common.JSONRef                `json:"entrypoint" yaml:"entrypoint"`
}
