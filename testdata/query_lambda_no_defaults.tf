resource rockset_query_lambda test {
  workspace = "commons"
  name      = "tpat-ql-diff"
  description = "basic lambda"
  sql {
    query = "{{ .Query }}"
  }
}

resource rockset_query_lambda_tag test {
  name = "test"
  workspace = rockset_query_lambda.test.workspace
  query_lambda = rockset_query_lambda.test.name
  version = rockset_query_lambda.test.version
}
