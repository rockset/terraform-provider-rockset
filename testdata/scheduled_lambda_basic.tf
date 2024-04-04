resource rockset_query_lambda "test_query_lambda" {
  workspace = "acc"
  name      = "test_query_lambda_name"
  sql {
    query = "select 1"
  }
}

resource "rockset_scheduled_lambda" "test_scheduled_lambda" {
  workspace = "acc"
  apikey = "var.ROCKSET_APIKEY"
  cron_string = "{{ .CronString }}"
  ql_name = rockset_query_lambda.test_query_lambda.name
  tag = "latest"
  total_times_to_execute = {{ .TotalTimesToExecute }}
}