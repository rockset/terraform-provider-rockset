variable KAFKA_BROKER {
  type = string
}

resource rockset_kafka_integration test {
  name                 = "acc-kafka-integration"
  description          = "Terraform provider acceptance tests."
  wait_for_integration = false
  kafka_data_format    = "JSON"
  kafka_topic_names    = ["test_json"]
}

# this relies on having a kafka-connect running in distributed mode
resource "null_resource" "configure-kafka-connect" {
  provisioner "local-exec" {
    command = <<EOT
curl -i http://${var.KAFKA_BROKER}/connectors -H "Content-Type: application/json" -X POST \
-d '{
  "name": "kafka-test",
  "config": {
    "name": "kafka-test",
    "connector.class": "rockset.RocksetSinkConnector",
    "tasks.max": 10,
    "topics": "test_json",
    "rockset.task.threads": 5,
    "rockset.apiserver.url": "https://api.usw2a1.rockset.com",
    "rockset.integration.key": "${rockset_kafka_integration.test.connection_string}",
    "format": "JSON",
    "key.converter": "org.apache.kafka.connect.storage.StringConverter",
    "value.converter": "org.apache.kafka.connect.storage.StringConverter",
    "key.converter.schemas.enable": false,
    "value.converter.schemas.enable": false
  }
}'
EOT
  }
}

resource rockset_kafka_collection test {
  name           = "kafka-collection"
  workspace      = "acc"
  description    = "Terraform provider acceptance tests."
  retention_secs = 3600

  source {
    integration_name    = rockset_kafka_integration.test.name
    topic_name          = "test_json"
    use_v3 = false
  }
  wait_for_documents = 1
  depends_on = [null_resource.configure-kafka-connect]
}
