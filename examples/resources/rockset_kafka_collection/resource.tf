variable "bootstrap_servers" {
  description = "Confluent Cloud bootstrap servers."
}

variable "apikey" {
  description = "Confluent Cloud API key."
}

variable "secret" {
  description = "Confluent Cloud secret."
}

resource rockset_kafka_integration confluent {
  name = "confluent-cloud"
  description = "Integration to ingest documents from Confluent Cloud"
  use_v3            = true
  bootstrap_servers = var.bootstrap_servers
  security_config = {
    api_key = var.apikey
    secret  = var.secret
  }
}

resource rockset_workspace confluent {
  name = "confluent"
  description = "Collections from Confluent Cloud topics."
}

resource rockset_kafka_collection test {
  name           = "confluent-cloud-collection"
  workspace      = rockset_workspace.confluent.name
  description    = "Collection from a Confluent Cloud topic."
  retention_secs = 3600

  source {
    integration_name = rockset_kafka_integration.confluent.name
    use_v3           = true
    topic_name       = "test_json"
    offset_reset_policy = "EARLIEST"
  }
}
