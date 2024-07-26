resource rockset_azure_blob_storage_integration test {
  name = "{{ .Name }}"
  connection_string = "BlobEndpoint=https://a.blob.core.windows.net/;SharedAccessSignature=sv=2022-11-02&ss=x=co&sp=x=2024-07-28T02:59:32Z&st=2024-07-26T18:59:32Z&spr=https&sig=x"
}

resource rockset_workspace test {
  name        = "{{ .Workspace }}"
  description = "{{ .Description }}"
}

resource rockset_azure_blob_storage_collection test {
  name           = "{{ .Collection }}"
  workspace      = rockset_workspace.test.name
  description    = "{{ .Description }}"
  retention_secs = 3600

  source {
    integration_name = rockset_azure_blob_storage_integration.test.name
    container           = "sampledatasets"
    pattern          = "cities.csv"
    format           = "csv"
    csv {
      first_line_as_column_names = false
      column_names               = [
        "country",
        "city",
        "population",
        "visited"
      ]
      column_types = [
        "STRING",
        "STRING",
        "STRING",
        "STRING"
      ]
    }
  }

  source {
    integration_name = rockset_azure_blob_storage_integration.test.name
    container           = "sampledatasets"
    pattern          = "cities.xml"
    format           = "xml"
    xml {
      root_tag = "cities"
      encoding = "UTF-8"
      doc_tag  = "city"
    }
  }
}
