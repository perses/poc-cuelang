package chart

#chart: {
  show_legend?: bool
  lines: [...#line]
}

#line: {
  expr: string
}