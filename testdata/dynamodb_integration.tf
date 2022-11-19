resource rockset_dynamodb_integration test {
  name                  = "{{ .Name }}"
  description           = "{{ .Description }}"
  aws_role_arn          = "{{ .Role }}"
  s3_export_bucket_name = "{{ .Bucket }}"
}
