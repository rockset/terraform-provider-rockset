resource rockset_query_lambda test {
  workspace = "commons"
  name      = "terraform-provider-acceptance-tests-query-lambda-tag-test"
  description = "basic lambda"
  sql {
    query = "SELECT * FROM commons._events LIMIT 1"
  }
}

resource rockset_query_lambda_tag test {
  name = "terraform_latest"
  workspace = rockset_query_lambda.test.workspace
  query_lambda = rockset_query_lambda.test.name
  version = rockset_query_lambda.test.version
}
