package main

import (
	"github.com/perses/common/app"

	"rndwww.nce.amadeus.net/git/MCVT/qa-go-server/api"
)

func main() {
	serverAPI := api.NewServerAPI()
	runner := app.NewRunner().
		WithDefaultHTTPServer("poc-cuelang")

	runner.HTTPServerBuilder().APIRegistration(serverAPI)

	runner.Start()
}
