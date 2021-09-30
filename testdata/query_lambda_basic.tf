resource rockset_query_lambda test {
  workspace = "commons"
  name      = "terraform-provider-acceptance-tests-query-lambda-basic"
  description = "basic lambda"
  sql {
    query = "SELECT * FROM commons._events WHERE _events._event_time > :start AND _events._event_time < :end LIMIT 1"
    default_parameter {
      name  = "start"
      type  = "start"
      value = "2020-01-01T00:00:00.000000Z"
    }

    default_parameter {
      name  = "end"
      type  = "end"
      value = "2200-01-01T00:00:00.000000Z"
    }
  }
}