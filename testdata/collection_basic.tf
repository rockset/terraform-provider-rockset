resource rockset_collection test {
  name        		= "{{ .Name }}"
  workspace   		= "{{ .Workspace }}"
  description 		= "{{ .Description }}"
  retention_secs 	= "{{ .Retention }}"
{{ if .InsertOnly }}
  insert_only = true
{{ end }}
{{ if (ne .IngestTransformation "") }}
field_mapping_query = "{{ .IngestTransformation }}"
{{ end }}
}
