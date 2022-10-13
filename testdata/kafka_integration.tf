resource rockset_kafka_integration test {
  name                 = "acc-kafka-integration"
  description          = "Terraform provider acceptance tests."
  kafka_topic_names    = ["foo", "bar"]
  kafka_data_format    = "JSON"
  wait_for_integration = false
}
