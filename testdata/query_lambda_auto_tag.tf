resource rockset_query_lambda test {
  workspace = "acc"
  name      = "{{ .Name  }}"
  description = "{{ .Description }}"
  sql {
    query = "{{ .SQL }}"
  }
}

resource rockset_query_lambda_auto_tag test {
  template = "{{ .Tag }}"
  max_tags = 2
  workspace = rockset_query_lambda.test.workspace
  query_lambda = rockset_query_lambda.test.name
  version = rockset_query_lambda.test.version
}
