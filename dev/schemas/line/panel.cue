package line

import (
  mylib "custom.io/foo/chart"
)

#panel: {
  displayed_name: string
  kind: "LineChart"
  chart: mylib.#chart
}

#panel