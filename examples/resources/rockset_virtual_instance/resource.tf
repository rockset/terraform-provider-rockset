resource "rockset_virtual_instance" "query" {
  name = "query"
  description = "vi for executing query lambdas"
  size = "MEDIUM"
  remount_on_resume = true
}

resource "rockset_collection_mount" "patch" {
  virtual_instance_id = rockset_virtual_instance.query.id
  path = "commons.data"
}
