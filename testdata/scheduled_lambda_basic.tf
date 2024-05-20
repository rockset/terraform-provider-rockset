resource rockset_query_lambda "test_query_lambda" {
  workspace = "acc"
  name      = "{{ .QueryLambdaName }}"
  sql {
    query = "select 1"
  }
}

resource "rockset_api_key" "test" {
  name = "{{ .QueryLambdaName }}"
}

resource "rockset_scheduled_lambda" "test_scheduled_lambda" {
  workspace = "acc"
  apikey = rockset_api_key.test.key
  cron_string = "{{ .CronString }}"
  query_lambda_name = rockset_query_lambda.test_query_lambda.name
  tag = "latest"
  total_times_to_execute = {{ .TotalTimesToExecute }}
}
