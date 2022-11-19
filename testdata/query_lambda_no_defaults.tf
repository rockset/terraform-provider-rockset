resource rockset_query_lambda test {
  workspace = "acc"
  name      = "{{ .Name  }}"
  description = "{{ .Description }}"
  sql {
    query = "{{ .SQL }}"
  }
}

resource rockset_query_lambda_tag test {
  name = "{{ .Tag }}"
  workspace = rockset_query_lambda.test.workspace
  query_lambda = rockset_query_lambda.test.name
  version = rockset_query_lambda.test.version
}
