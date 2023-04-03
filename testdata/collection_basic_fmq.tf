resource rockset_collection test {
  name                = "{{ .Name }}"
  workspace           = "{{ .Workspace }}"
  description         = "{{ .Description }}"
  retention_secs      = "{{ .Retention }}"
  field_mapping_query = "{{ .IngestTransformation }}"
}
