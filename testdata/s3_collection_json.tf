resource rockset_s3_integration test {
  name = "terraform-provider-acceptance-tests-s3-collection-json"
  description = "Terraform provider acceptance tests."
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
}

resource rockset_s3_collection test {
  name              = "terraform-provider-acceptance-tests-s3-json"
  workspace         = "commons"
  description       = "Terraform provider acceptance tests."
  retention_secs    = 3600

  source {
    integration_name  = rockset_s3_integration.test.name
    bucket            = "terraform-provider-rockset-tests"
    pattern           = "cities.json"
    format            = "json"
  }

}