package line

import (
  // About "custom.io" : imports without a domain are assumed to be builtins from the standard library
  // You don’t actually need a domain or repository, it’s just a path naming requirement
  mylib "custom.io/foo/chart"
)

#panel: {
  displayed_name: string
  kind: "LineChart"
  chart: mylib.#chart
}

#panel