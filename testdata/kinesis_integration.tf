resource rockset_kinesis_integration test {
  name = "terraform-provider-acceptance-test-kinesis-integration"
  description = "Terraform provider acceptance tests."
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests-kinesis"
}