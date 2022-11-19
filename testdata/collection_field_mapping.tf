resource rockset_collection test {
  name           = "{{ .Name }}"
  workspace      = "{{ .Workspace }}"
  description    = "{{ .Description }}"
  retention_secs = {{ .Retention }}

  field_mapping {
    name = "string to float"
    input_fields {
      field_name = "population"
      if_missing = "SKIP"
      is_drop    = false
      param      = "pop"
    }

    output_field {
      field_name = "pop"
      on_error   = "FAIL"
      sql        = "CAST(:pop as int)"
    }
  }
}