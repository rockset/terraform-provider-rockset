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
  name              = "terraform-provider-acceptance-test-kafka-integration"
  description       = "Terraform provider acceptance tests."
  use_v3            = true
  bootstrap_servers = var.CC_BOOTSTRAP_SERVERS
  security_config = {
    api_key = var.CC_KEY
    secret  = var.CC_SECRET
  }
}
