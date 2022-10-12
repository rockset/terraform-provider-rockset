variable CC_BOOTSTRAP_SERVERS {
  type = string
}

variable CC_KEY {
  type = string
}

variable CC_SECRET {
  type = string
}

resource rockset_kafka_integration test {
  name              = "acc-kafka-integration"
  description       = "Terraform provider acceptance tests."
  use_v3            = true
  bootstrap_servers = var.CC_BOOTSTRAP_SERVERS
  security_config   = {
    api_key = var.CC_KEY
    secret  = var.CC_SECRET
  }
}

resource rockset_kafka_collection test {
  name           = "kafka-collection"
  workspace      = "acc"
  description    = "Terraform provider acceptance tests."
  retention_secs = 3600

  source {
    integration_name = rockset_kafka_integration.test.name
    use_v3           = true
    topic_name       = "test_json"
    offset_reset_policy = "EARLIEST"
  }
}
