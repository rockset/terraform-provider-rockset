resource rockset_s3_integration test {
  name = "terraform-provider-acceptance-tests-s3-xml"
  description = "Terraform provider acceptance tests."
  aws_role_arn = "arn:aws:iam::469279130686:role/terraform-provider-rockset-tests"
}

resource rockset_s3_collection test {
  name              = "terraform-provider-acceptance-tests-s3-xml"
  workspace         = "commons"
  description       = "Terraform provider acceptance tests."
  retention_secs    = 3600
  integration_name  = rockset_s3_integration.test.name
  bucket            = "terraform-provider-rockset-tests"
  pattern           = "note.xml"
  format            = "xml"
  xml {
    root_tag = "note"
    encoding = "UTF-8"
    doc_tag  = "note"
  }
}