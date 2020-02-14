data "rockset_account" "demo" {}

resource "rockset_s3" "demo" {
  name = "provider integration"
  aws_role_arn = module.rockset.rockset-role-arn
}

module "rockset" {
  source = "rockset/s3/integration"
  bucket = "rockset-pme-provider-test-bucket"
  rockset-role-name = "provider-integration-test"
  rockset-account-id = data.rockset_account.demo.account_id
  rockset-external-id = "your AWS external id here"
}
