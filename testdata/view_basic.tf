resource rockset_workspace test {
  name        = "{{ .Workspace }}"
  description = "{{ .Description }}"
}

resource rockset_view test {
  workspace   = rockset_workspace.test.name
  name        = "{{ .Name }}"
  query       = "{{ .SQL }}"
  description = "{{ .Description }}"
}
