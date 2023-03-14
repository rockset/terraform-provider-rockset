resource rockset_workspace sample {
  name        = "sample"
  description = "sample datasets"
}

resource rockset_s3_integration public {
  name         = "rockset-public-collections"
  description  = "Integration to access Rockset's public datasets"
  aws_role_arn = "arn:aws:iam::469279130686:role/rockset-public-datasets"
}

resource rockset_s3_collection cities {
  workspace = rockset_workspace.sample.name
  name      = "cities"

  source {
    bucket           = "rockset-public-datasets"
    integration_name = rockset_s3_integration.public.name
    pattern          = "cities/*.json"
    format           = "json"
  }
}
