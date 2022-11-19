resource rockset_collection test {
  name           = "{{ .Name }}"
  workspace      = "{{ .Workspace }}"
  description    = "{{ .Description }}"
  retention_secs = {{ .Retention }}

  clustering_key {
    field_name = "population"
    type       = "AUTO"
  }
}
