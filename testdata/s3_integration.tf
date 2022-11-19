resource rockset_s3_integration test {
  name = "{{ .Name }}"
  description = "{{ .Description }}"
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
}