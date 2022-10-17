resource rockset_view test {
  workspace   = "acc"
  name        = "{{ .Name }}"
  query       = "select * from commons._events where _events.kind = 'COLLECTION'"
  description	= "{{ .Description }}"
}
