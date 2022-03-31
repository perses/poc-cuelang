// Copyright 2021 The Perses Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"

	"github.com/perses/common/app"
	"github.com/sirupsen/logrus"

	"github.com/perses/poc-cuelang/api"
	"github.com/perses/poc-cuelang/internal/config"
	"github.com/perses/poc-cuelang/internal/validator"
)

func main() {
	configFile := flag.String("config", "", "Path to the yaml configuration file. Configuration can be overridden with environment variables.")
	flag.Parse()

	conf, err := config.Resolve(*configFile)
	if err != nil {
		logrus.WithError(err).Fatalf("error reading configuration file %q", *configFile)
	}

	validator := validator.New(conf)
	serverAPI := api.NewServerAPI(validator)
	runner := app.NewRunner().WithDefaultHTTPServer("poc_cuelang")
	runner.HTTPServerBuilder().APIRegistration(serverAPI)
	runner.Start()
}
