resource "rockset_virtual_instance" "test" {
  name = "{{ .Name }}"
  description = "{{ .Description }}"
  size = "{{ .Size }}"
  remount_on_resume = {{ .Remount }}
}

resource "rockset_collection_mount" "patch" {
  virtual_instance_id = rockset_virtual_instance.test.id
  path = "persistent.patch"
}
