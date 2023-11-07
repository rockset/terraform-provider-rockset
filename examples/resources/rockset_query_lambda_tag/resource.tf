resource "rockset_query_lambda_tag" "active" {
  name         = "active"
  query_lambda = "top-movies"
  workspace    = "commons"
  version      = "b22fb578b8106694"
}
