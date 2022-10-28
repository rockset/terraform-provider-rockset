resource rockset_workspace sample {
  name = "sample"
  description = "sample datasets"
}

resource rockset_sample_collection cities {
  workspace = rockset_workspace.sample.name
  name      = "cities"
  dataset   = "cities"
}
