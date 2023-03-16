resource rockset_dynamodb_integration test {
  name                  = "{{ .Name }}"
  description           = "{{ .Description }}"
  aws_role_arn          = "{{ .Role }}"
  s3_export_bucket_name = "{{ .Bucket }}"
}

resource rockset_dynamodb_collection test {
  name           = "{{ .Collection }}"
  workspace      = "{{ .Workspace }}"
  description    = "{{ .Description }}"
  retention_secs = 3600


  source {
    integration_name = rockset_dynamodb_integration.test.name
    table_name       = "terraform-provider-rockset-tests-1"
    aws_region       = "us-west-2"
{{ if ( .UseScanApi ) }}
    use_scan_api = "{{ .UseScanApi }}"
{{ end }}
{{ if (.RCU ) }}
    rcu = "{{ .RCU }}"
{{ end }}
  }

  source {
    integration_name = rockset_dynamodb_integration.test.name
    table_name       = "terraform-provider-rockset-tests-2"
    aws_region       = "us-west-2"
{{ if ( .UseScanApi ) }}
    use_scan_api = "{{ .UseScanApi }}"
{{ end }}
{{ if ( .RCU ) }}
    rcu = "{{ .RCU }}"
{{ end }}
  }
}
