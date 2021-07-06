variable GCS_SERVICE_ACCOUNT_KEY {
  type = string
}

resource rockset_gcs_integration test {
  name = "terraform-provider-acceptance-test-gcs-integration"
  description = "Terraform provider acceptance tests."
  service_account_key = var.GCS_SERVICE_ACCOUNT_KEY
}
