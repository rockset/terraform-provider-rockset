resource rockset_s3_integration test {
  name         = "{{ .Name }}"
  description  = "{{ .Description }}"
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
}

resource rockset_workspace test {
  name        = "{{ .Workspace }}"
  description = "{{ .Description }}"
}

resource rockset_s3_collection test {
  name           = "{{ .Collection }}"
  workspace      = rockset_workspace.test.name
  description    = "{{ .Description }}"
  retention_secs = 3600

  source {
    integration_name = rockset_s3_integration.test.name
    bucket           = "terraform-provider-rockset-tests"
    pattern          = "cities.json"
    format           = "json"
  }
}
