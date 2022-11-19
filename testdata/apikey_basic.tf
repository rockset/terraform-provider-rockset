resource rockset_api_key test {
  name = "{{ .Name }}"
  role = "read-only"
}
