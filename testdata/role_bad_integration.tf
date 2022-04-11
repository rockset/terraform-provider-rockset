resource rockset_role test {
  name        = "terraform-provider-acceptance-tests"
  description = "Terraform provider acceptance tests"
  privilege {
    action        = "CREATE_COLLECTION_INTEGRATION"
    resource_name = "dummy"
    cluster       = "dummy"
  }
}
