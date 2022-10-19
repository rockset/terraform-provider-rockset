resource rockset_workspace sample {
  name = "sample"
  description = "sample datasets"
}

resource rockset_s3_integration public {
  name = "rockset-public-collections"
  description = "Integration to access Rockset's public datasets"
  aws_role_arn = "arn:aws:iam::469279130686:role/rockset-public-datasets"
}

resource rockset_s3_collection cities {
  workspace = rockset_workspace.sample.name
  name = "cities"
  integration_name = rockset_s3_integration.public.name
  source = {
    bucket = "rockset-public-datasets"
    pattern = "cities/*.json"
    format = "json"
  }
}
