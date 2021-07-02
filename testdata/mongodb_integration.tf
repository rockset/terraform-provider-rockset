variable MONGODB_CONNECTION_URI {
  type = string
}
resource rockset_mongodb_integration test {
  name = "terraform-provider-acceptance-test-mongodb-integration"
  description = "Terraform provider acceptance tests."
  connection_uri = var.MONGODB_CONNECTION_URI
}