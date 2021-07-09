resource rockset_kinesis_integration test {
  name = "terraform-provider-acceptance-test-kinesis-collection"
  description = "Terraform provider acceptance tests."
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-kinesis"
}

resource rockset_kinesis_collection test {
  name              = "terraform-provider-acceptance-tests-kinesis"
  workspace         = "commons"
  description       = "Terraform provider acceptance tests."
  retention_secs    = 3600

  source {
    integration_name  = rockset_kinesis_integration.test.name
    stream_name = "terraform-provider-rockset-tests-kinesis"
    format = "json"
    aws_region = "us-west-2"
  }

  source {
    integration_name  = rockset_kinesis_integration.test.name
    stream_name = "terraform-provider-rockset-tests-kinesis"
    format = "mysql"
    dms_primary_key = [
      "foo",
      "bar"
    ]
  }

  source {
    integration_name  = rockset_kinesis_integration.test.name
    stream_name = "terraform-provider-rockset-tests-kinesis"
    format = "postgres"
    dms_primary_key = [
      "foo",
      "bar"
    ]
  }
}