---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_kafka_integration Resource - terraform-provider-rockset"
subcategory: ""
description: |-
  Manages a Rockset Kafka Integration.
  If the integration is connected with Confluent Cloud, there is a Terraform provider https://registry.terraform.io/providers/confluentinc/confluent/latest/docs which can be used to configure the Confluent Cloud side of the integration.
---

# rockset_kafka_integration (Resource)

Manages a Rockset Kafka Integration.

If the integration is connected with Confluent Cloud, there is a [Terraform provider](https://registry.terraform.io/providers/confluentinc/confluent/latest/docs) which can be used to configure the Confluent Cloud side of the integration.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique identifier for the integration. Can contain alphanumeric or dash characters.

### Optional

- `bootstrap_servers` (String) The Kafka bootstrap server url(s). Required only for V3 integration.
- `connection_string` (String) Kafka connection string.
- `description` (String) Text describing the integration.
- `id` (String) The ID of this resource.
- `kafka_data_format` (String) The format of the Kafka topics being tailed. Possible values: JSON, AVRO.
- `kafka_topic_names` (Set of String) Kafka topics to tail.
- `schema_registry_config` (Map of String) Kafka configuration for schema registry. Required only for V3 integration.
- `security_config` (Map of String) Kafka security configurations. Required only for V3 integration.
- `use_v3` (Boolean) Use v3 for Confluent Cloud.
- `wait_for_integration` (Boolean) Wait until the integration is active.

