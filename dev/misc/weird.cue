#piechart: {
  displayed_name: string,
  kind: "PieChart",
  chart: {
    queries: [...#query]
  }
}

#query: {
  expr: string
  legend?: string
  weird: bool
}

#piechart