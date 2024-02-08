data "rockset_virtual_instance" "main" {
  id = "29e4a43c-fff4-4fe6-80e3-1ee57bc22e82"
}

resource "rockset_autoscaling_policy" "main" {
  virtual_instance_id = data.rockset_virtual_instance.main.id
  enabled = true
  min_size = "MEDIUM"
  max_size = "XLARGE2"
}
