resource rockset_dynamodb_integration test {
	name = "terraform-provider-acceptance-test-dynamodb-collection-1"
	description = "Terraform provider acceptance tests."
	aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo"
}

resource rockset_dynamodb_collection test {
  name              = "terraform-provider-acceptance-tests-dynamodb"
  workspace         = "commons"
  description       = "Terraform provider acceptance tests."
  retention_secs    = 3600

  source {
    integration_name  = rockset_dynamodb_integration.test.name
    table_name = "terraform-provider-rockset-tests-1"
    aws_region = "us-west-2"
    rcu = 5
  }

  source {
    integration_name  = rockset_dynamodb_integration.test.name
    table_name = "terraform-provider-rockset-tests-2"
    aws_region = "us-west-2"
    rcu = 5
  }

}