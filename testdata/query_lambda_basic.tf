resource rockset_query_lambda test {
  workspace = "commons"
  name      = "terraform-provider-acceptance-tests-query-lambda-basic"
  description = "basic lambda"
  sql {
    query = "SELECT * FROM commons._events WHERE _events._event_time > :timestamp LIMIT 1"
    default_parameter {
      name  = "timestamp"
      type  = "timestamp"
      value = "2020-01-01T00:00:00.000000Z"
    }
  }
}