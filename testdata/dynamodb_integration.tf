resource rockset_dynamodb_integration test {
  name = "terraform-provider-acceptance-test-dynamodb-integration"
  description = "Terraform provider acceptance tests."
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-dynamo"
  s3_export_bucket_name = "terraform-provider-rockset-tests"
}
