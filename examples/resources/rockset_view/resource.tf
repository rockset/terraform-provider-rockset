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

resource "rockset_view" "brazil" {
  name      = "brazil"
  query     = "SELECT * FROM sample.cities c WHERE c.fields.country_code = 'BR'"
  workspace = rockset_workspace.sample.name
  depends_on = [rockset_s3_collection.cities]
}
