package pie

#panel: {
  displayed_name: string,
  kind: "PieChart",
  chart: {
    queries: [...#query]
  }
}

#panel