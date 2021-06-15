output curl_command {
  value = "curl -s -X POST https://$ROCKSET_APISERVER/v1/orgs/self/ws/${rockset_workspace.example.name}/lambdas/${rockset_query_lambda.example.name}/versions/${rockset_query_lambda.example.version} -H \"Authorization: ApiKey $ROCKSET_APIKEY\""
}
