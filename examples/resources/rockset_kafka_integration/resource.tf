variable "rockset_apiserver" {
  description = ""
  default = "https://api.usw2a1.rockset.com"
}

variable "kafka_connect" {
  description = "Host and port of your local kafka-connect"
  default = "localhost:8083"
}

variable "topics" {
  description = "Kafka topics to ingest from."
}

variable "max_tasks" {
  default = 10
}

resource rockset_kafka_integration local {
  name                 = "local-kafka-integration"
  description          = "Integration to ingest from a local kafka."
  kafka_topic_names    = ["foo", "bar"]
  kafka_data_format    = "JSON"
  wait_for_integration = false
}

resource "null_resource" "configure-kafka-connect" {
  provisioner "local-exec" {
    command = <<EOT
curl -i http://${var.kafka_connect}/connectors -H "Content-Type: application/json" -X POST \
-d '{
  "name": "kafka-test",
  "config": {
    "name": "kafka-test",
    "connector.class": "rockset.RocksetSinkConnector",
    "tasks.max": ${var.max_tasks},
    "topics": "${var.topics}",
    "rockset.task.threads": 5,
    "rockset.apiserver.url": "${var.rockset_apiserver}",
    "rockset.integration.key": "${rockset_kafka_integration.local.connection_string}",
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
