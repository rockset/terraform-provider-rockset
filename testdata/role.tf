resource rockset_role test {
  name        = "terraform-provider-acceptance-tests"
  description = "Terraform provider acceptance tests"
  privilege {
    action = "GET_METRICS_GLOBAL"
  }
  privilege {
    action        = "CREATE_COLLECTION_INTEGRATION"
    resource_name = "dummy"
  }
  privilege {
    action        = "QUERY_DATA_WS"
    resource_name = "common"
    cluster       = "*ALL*"
  }
  privilege {
    action        = "EXECUTE_QUERY_LAMBDA_WS"
    resource_name = "common"
    cluster       = "usw2a1"
  }
  privilege {
    action        = "QUERY_VI"
    resource_name = "29e4a43c-fff4-4fe6-80e3-1ee57bc22e82"
    cluster       = "usw2a1"
  }
  privilege {
    action        = "LIST_RESOURCES_WS"
    resource_name = "common"
    cluster = "*ALL*"
    // TODO: cluster should default to *ALL* when not specify
  }
}
