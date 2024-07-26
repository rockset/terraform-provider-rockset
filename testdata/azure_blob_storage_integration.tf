resource rockset_azure_blob_storage_integration test {
  name = "{{ .Name }}"
  connection_string = "BlobEndpoint=https://a.blob.core.windows.net/;SharedAccessSignature=sv=2022-11-02&ss=x=co&sp=x=2024-07-28T02:59:32Z&st=2024-07-26T18:59:32Z&spr=https&sig=x"
}
