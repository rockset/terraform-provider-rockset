resource "rockset_virtual_instance" "test" {
  name = "{{ .Name }}"
  description = "{{ .Description }}"
  size = "{{ .Size }}"
  remount_on_resume = {{ .Remount }}
  auto_suspend_seconds = 900
}

resource "rockset_collection_mount" "patch" {
  virtual_instance_id = rockset_virtual_instance.test.id
  path = "persistent.patch"
}
