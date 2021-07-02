variable MONGODB_CONNECTION_URI {
	type = string
}

resource rockset_mongodb_integration test {
	name = "terraform-provider-acceptance-test-mongodb-collection"
	description = "Terraform provider acceptance tests."
	connection_uri = var.MONGODB_CONNECTION_URI
}

resource rockset_mongodb_collection test {
  name              = "terraform-provider-acceptance-tests-mongodb"
  workspace         = "commons"
  description       = "Terraform provider acceptance tests."
  retention_secs    = 3600

  source {
    integration_name  = rockset_mongodb_integration.test.name
    database_name = "sample_analytics"
    collection_name = "accounts"
  }

	source {
    integration_name  = rockset_mongodb_integration.test.name
    database_name = "sample_analytics"
    collection_name = "customers"
  }
}