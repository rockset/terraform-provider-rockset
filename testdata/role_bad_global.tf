resource rockset_role test {
  name        = "terraform-provider-acceptance-tests"
  description = "Terraform provider acceptance tests"
  privilege {
    action        = "UPDATE_VI_GLOBAL"
    resource_name = "dummy"
  }
}
