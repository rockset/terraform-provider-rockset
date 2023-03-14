resource rockset_workspace test {
  name        = "{{ .Workspace }}"
  description = "{{ .Workspace }}"
}

resource rockset_alias test {
  name        = "{{ .Name }}"
  description = "{{ .Description }}"
  workspace   = rockset_workspace.test.name
  collections = ["{{ .Alias }}"]
}
