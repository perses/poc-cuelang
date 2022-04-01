#linechart: {
  displayed_name: string
  kind: "LineChart"
  chart: #chart
}

#chart: {
  show_legend?: bool
  lines: [...#line]
}

#line: {
  expr: string
}

#linechart