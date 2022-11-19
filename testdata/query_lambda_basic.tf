resource rockset_query_lambda test {
  workspace = "acc"
  name      = "{{ .Name }}"
  description = "{{ .Description }}"
  sql {
    query = "{{ .SQL }}"
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