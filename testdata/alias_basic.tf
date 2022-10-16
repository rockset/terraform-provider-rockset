resource rockset_alias test {
  name        = "{{ .Name }}"
  description	= "{{ .Description }}"
  workspace		= "acc"
  collections = ["{{ .Alias }}"]
}
