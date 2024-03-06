resource rockset_collection test {
  name        		= "{{ .Name }}"
  workspace   		= "{{ .Workspace }}"
  description 		= "{{ .Description }}"
{{ if (ne .Retention -1) }}
  retention_secs 	= "{{ .Retention }}"
{{ end }}
{{ if (ne .StorageCompressionType "") }}
  storage_compression_type 	= "{{ .StorageCompressionType }}"
{{ end }}
{{ if (ne .IngestTransformation "") }}
  ingest_transformation = "{{ .IngestTransformation }}"
{{ end }}
{{ if (ne .CreateTimeout "") }}
  timeouts {
        create = "{{.CreateTimeout}}"
  }
{{ end }}
}
