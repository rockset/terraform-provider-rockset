resource rockset_collection test {
  name        		= "{{ .Name }}"
  workspace   		= "{{ .Workspace }}"
  description 		= "{{ .Description }}"
{{ if (ne .Retention -1) }}
  retention_secs 	= "{{ .Retention }}"
{{ end }}
{{ if (ne .IngestTransformation "") }}
  field_mapping_query = "{{ .IngestTransformation }}"
{{ end }}
{{ if (ne .CreateTimeout "") }}
  timeouts {
        create = "{{.CreateTimeout}}"
  }
{{ end }}
}
