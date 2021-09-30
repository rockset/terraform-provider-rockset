resource rockset_query_lambda test {
  workspace = "commons"
  name      = "terraform-provider-acceptance-tests-query-lambda-no-defaults"
  description = "basic lambda"
  sql {
    query = "SELECT * FROM commons._events LIMIT 1"
  }
}