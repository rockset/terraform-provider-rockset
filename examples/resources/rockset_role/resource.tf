resource "rockset_role" "query-only" {
  name        = "query-only"
  description = "This role can only query collections in the prod workspace in the usw2a1 cluster"

  privilege {
    action        = "QUERY_DATA_WS"
    resource_name = "prod"
    cluster       = "usw2a1"
  }
  privilege {
    action        = "EXECUTE_QUERY_LAMBDA_WS"
    resource_name = "prod"
    cluster       = "usw2a1"
  }
}

resource "rockset_api_key" "query-only" {
  name = "query-only"
  role = rockset_role.query-only.name
}
