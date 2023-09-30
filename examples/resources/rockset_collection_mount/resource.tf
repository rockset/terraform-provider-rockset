resource "rockset_collection_mount" "data" {
  virtual_instance_id = rockset_virtual_instance.secondary.id
  path = "commons.data"
}
